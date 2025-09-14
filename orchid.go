// Package orchid provides a colorized, structured logging library for Go applications.
//
// Orchid supports different severity levels (INFO, OK, WARN, ERROR, FATAL, DEBUG)
// with ANSI color-coded console output and optional file logging in both text
// and JSON formats. The library uses a global configuration system for managing
// logging settings including colors, file output, and formatting.
//
// Basic usage with the default logger:
//
//	orchid.Init("my-app")
//	orchid.Info("Application starting")
//	orchid.Error("Something went wrong")
//
// Usage with file logging:
//
//	orchid.Init("my-app")
//	err := orchid.SetLogFile("app.log", orchid.FormatJSON)
//	if err != nil {
//		log.Fatal(err)
//	}
//	orchid.Info("This will be logged to both console and file")
//
// Usage with custom logger instances:
//
//	// Set global file for all loggers
//	err := orchid.SetLogFile("app.log", orchid.FormatTXT)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Create individual loggers - they all write to the same global file
//	var dbLogger orchid.Logger
//	err = dbLogger.Init("database")
//	if err != nil {
//		log.Fatal(err)
//	}
//	dbLogger.Info("Database connection established")
//
//	var apiLogger orchid.Logger
//	err = apiLogger.Init("api")
//	if err != nil {
//		log.Fatal(err)
//	}
//	apiLogger.Info("API server started")
//
// Global configuration can be used to control logging behavior:
//
//	config := orchid.GetConfiguration()
//	config.SetEnableColors(false) // Disable color output
//	config.SetDefaultFile("global.log")
//	config.SetDefaultFormat(orchid.FormatJSON)
//
// For proper resource cleanup, especially when using file logging:
//
//	defer orchid.Close() // Clean up file handles before program exit
package orchid

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// ANSI color codes for different log levels
const (
	COLOR_RESET = "\033[0m"       // Reset color
	COLOR_INFO  = "\033[48;5;33m" // Blue background for INFO
	COLOR_OK    = "\033[48;5;36m" // Cyan background for OK
	COLOR_WARN  = "\033[48;5;3m"  // Yellow background for WARN
	COLOR_ERROR = "\033[48;5;1m"  // Red background for ERROR
	COLOR_FATAL = "\033[48;5;1m"  // Red background for FATAL
	COLOR_DEBUG = "\033[48;5;5m"  // Magenta background for DEBUG
)

// FileFormat represents the format for file logging output.
type FileFormat int

// Available file formats for logging output.
const (
	FormatTXT  FileFormat = iota // Plain text format
	FormatJSON                   // JSON format
)

// logMessage represents an internal log message structure.
type logMessage struct {
	Severity string    // Log severity level (INFO, ERROR, etc.)
	Text     string    // Log message text
	Module   string    // Module name that generated the log
	Time     time.Time // Timestamp when the log was created
}

// Logger represents a structured logger instance for a specific module.
// Each Logger instance is associated with a module name and writes to both
// console output and the global log file (if configured).
// All Logger instances share the same global file configuration.
// Logger is safe for concurrent use by multiple goroutines.
type Logger struct {
	mu     sync.Mutex // Protects all fields and operations
	module string     // Module name for this logger instance
}

// Init initializes the Logger with a module name.
// The logger will write to both console output and the global log file
// (if one is configured via SetLogFile).
// Returns an error if the module name is invalid.
func (l *Logger) Init(moduleName string) error {
	// Validate module name before acquiring lock
	trimmed := strings.TrimSpace(moduleName)
	if trimmed == "" {
		return fmt.Errorf("module name cannot be empty or whitespace-only")
	}
	if len(trimmed) > 50 {
		return fmt.Errorf("module name too long (max 50 characters): %d", len(trimmed))
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	l.module = trimmed

	return nil
}

// createLogMessage creates a logMessage struct with the given severity and message.
func (l *Logger) createLogMessage(severity string, a ...interface{}) logMessage {
	return logMessage{
		Severity: severity,
		Text:     fmt.Sprint(a...),
		Module:   l.module,
		Time:     time.Now(),
	}
}

// writeToFile writes a log message to the file in the specified format.
// Returns an error if the write operation fails.
func (l *Logger) writeToFile(msg logMessage) error {
	config := GetConfiguration()
	if config.GetDefaultFile() == "" {
		return nil // No file configured - not an error
	}

	logFile := config.getLogFile()
	if logFile == nil {
		return fmt.Errorf("log file configured but file handle is not available")
	}

	switch config.GetDefaultFormat() {
	case FormatTXT:
		txtMessage := fmt.Sprintf("%s [%s] %s: %s",
			msg.Time.Format("2006-01-02 15:04:05"),
			msg.Severity,
			msg.Module,
			msg.Text)
		if _, err := fmt.Fprintln(logFile, txtMessage); err != nil {
			return fmt.Errorf("failed to write text log to file: %w", err)
		}
	case FormatJSON:
		jsonData, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("failed to marshal log message to JSON: %w", err)
		}
		if _, err := fmt.Fprintln(logFile, string(jsonData)); err != nil {
			return fmt.Errorf("failed to write JSON log to file: %w", err)
		}
	default:
		return fmt.Errorf("unsupported log format: %d", config.GetDefaultFormat())
	}

	return nil
}

// printLogMessage outputs a log message to the console with colors and optionally to file.
// FATAL messages will call log.Fatal() which exits the program.
func (l *Logger) printLogMessage(msg logMessage) {
	metadata := fmt.Sprintf("%-20s %-6s", msg.Module, msg.Severity)

	config := GetConfiguration()
	var consoleMessage string

	if config.GetEnableColors() {
		color := COLOR_INFO
		switch msg.Severity {
		case "INFO":
			color = COLOR_INFO
		case "OK":
			color = COLOR_OK
		case "WARN":
			color = COLOR_WARN
		case "ERROR":
			color = COLOR_ERROR
		case "FATAL":
			color = COLOR_FATAL
		case "DEBUG":
			color = COLOR_DEBUG
		}
		consoleMessage = fmt.Sprintf("%s %s %s %s %s", COLOR_RESET, color, metadata, COLOR_RESET, msg.Text)
	} else {
		// No colors - just plain text
		consoleMessage = fmt.Sprintf("%s %s", metadata, msg.Text)
	}

	// Attempt to write to file, but don't let file errors prevent console logging
	if err := l.writeToFile(msg); err != nil {
		// Log file write errors to stderr without breaking the logging flow
		fmt.Fprintf(os.Stderr, "ORCHID FILE ERROR: %v\n", err)
	}

	if msg.Severity == "FATAL" {
		log.Fatal(consoleMessage)
	} else {
		log.Println(consoleMessage)
	}
}

// log is the internal method that handles logging for all severity levels.
func (l *Logger) log(severity string, a ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	msg := l.createLogMessage(severity, a...)
	l.printLogMessage(msg)
}

// Info logs a message at INFO level with blue background color.
func (l *Logger) Info(a ...interface{}) {
	l.log("INFO", a...)
}

// OK logs a message at OK level with cyan background color.
func (l *Logger) OK(a ...interface{}) {
	l.log("OK", a...)
}

// Error logs a message at ERROR level with red background color.
func (l *Logger) Error(a ...interface{}) {
	l.log("ERROR", a...)
}

// Fatal logs a message at FATAL level with red background color and exits the program.
func (l *Logger) Fatal(a ...interface{}) {
	l.log("FATAL", a...)
}

// Warn logs a message at WARN level with yellow background color.
func (l *Logger) Warn(a ...interface{}) {
	l.log("WARN", a...)
}

// Debug logs a message at DEBUG level with magenta background color.
func (l *Logger) Debug(a ...interface{}) {
	l.log("DEBUG", a...)
}

var (
	defaultLogger Logger
	defaultMu     sync.Mutex // Protects defaultLogger initialization
)

// Init initializes the default logger with console-only output.
// This is a convenience function for simple logging without file output.
// Returns an error if the module name is invalid.
func Init(moduleName string) error {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	return defaultLogger.Init(moduleName)
}

// Info logs a message at INFO level using the default logger.
func Info(a ...interface{}) {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	defaultLogger.log("INFO", a...)
}

// OK logs a message at OK level using the default logger.
func OK(a ...interface{}) {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	defaultLogger.log("OK", a...)
}

// Error logs a message at ERROR level using the default logger.
func Error(a ...interface{}) {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	defaultLogger.log("ERROR", a...)
}

// Fatal logs a message at FATAL level using the default logger and exits the program.
func Fatal(a ...interface{}) {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	defaultLogger.log("FATAL", a...)
}

// Warn logs a message at WARN level using the default logger.
func Warn(a ...interface{}) {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	defaultLogger.log("WARN", a...)
}

// Debug logs a message at DEBUG level using the default logger.
func Debug(a ...interface{}) {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	defaultLogger.log("DEBUG", a...)
}

// SetLogFile sets the global log file and format for ALL loggers.
// This affects both the default logger and all individual Logger instances.
// All loggers will write to the same file using the specified format.
func SetLogFile(filePath string, format FileFormat) error {
	// Validate inputs before acquiring lock
	if format < FormatTXT || format > FormatJSON {
		return fmt.Errorf("invalid log format: %d (must be between %d and %d)", format, FormatTXT, FormatJSON)
	}

	// Allow empty filePath to disable file logging
	if filePath != "" {
		// Basic file path validation
		if strings.TrimSpace(filePath) != filePath {
			return fmt.Errorf("file path cannot have leading or trailing whitespace")
		}
		if strings.Contains(filePath, "\x00") {
			return fmt.Errorf("file path cannot contain null bytes")
		}
		if len(filePath) > 260 {
			return fmt.Errorf("file path too long (max 260 characters): %d", len(filePath))
		}
	}

	defaultMu.Lock()
	defer defaultMu.Unlock()

	config := GetConfiguration()
	err := config.SetDefaultFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to set log file: %w", err)
	}
	config.SetDefaultFormat(format)

	return nil
}

// Close closes any open file handles in the global configuration.
// This should be called before program exit to ensure proper cleanup.
func Close() error {
	config := GetConfiguration()
	return config.Close()
}
