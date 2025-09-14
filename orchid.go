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
//	var logger orchid.Logger
//	err := logger.Init("database")
//	if err != nil {
//		log.Fatal(err)
//	}
//	err = logger.SetLogFile("db.log", orchid.FormatTXT)
//	if err != nil {
//		log.Fatal(err)
//	}
//	logger.Info("Database connection established")
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

// Logger represents a structured logger instance with optional file output.
// Each Logger instance is associated with a specific module name and can
// optionally write to a file in addition to console output.
// Logger is safe for concurrent use by multiple goroutines.
type Logger struct {
	mu     sync.Mutex // Protects all fields and operations
	module string     // Module name for this logger instance
}

// Init initializes the Logger with a module name, optional file path, and file format.
// If filePath is empty, only console logging will be used.
// If filePath is provided, logs will be written to both console and file.
// If the logger already has a file open, it will be closed before opening the new one.
// Returns an error if the file cannot be opened for writing.
func (l *Logger) Init(moduleName string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.module = moduleName

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
func (l *Logger) writeToFile(msg logMessage) {
	config := GetConfiguration()
	if config.GetDefaultFile() == "" {
		return // No file configured
	}

	logFile := config.getLogFile()
	if logFile == nil {
		return // No file handle available
	}

	switch config.GetDefaultFormat() {
	case FormatTXT:
		txtMessage := fmt.Sprintf("%s [%s] %s: %s",
			msg.Time.Format("2006-01-02 15:04:05"),
			msg.Severity,
			msg.Module,
			msg.Text)
		fmt.Fprintln(logFile, txtMessage)
	case FormatJSON:
		jsonData, err := json.Marshal(msg)
		if err != nil {
			fmt.Fprintf(logFile, "Error marshaling log message to JSON: %v\n", err)
			return
		}
		fmt.Fprintln(logFile, string(jsonData))
	}
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

	l.writeToFile(msg)

	if msg.Severity == "FATAL" {
		log.Fatal(consoleMessage)
	} else {
		log.Println(consoleMessage)
	}
}

func (l *Logger) SetLogFile(filePath string, format FileFormat) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	config := GetConfiguration()
	err := config.SetDefaultFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to set log file: %v", err)
	}
	config.SetDefaultFormat(format)

	return nil
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
func Init(moduleName string) {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	defaultLogger.Init(moduleName)
}

// Info logs a message at INFO level using the default logger.
func Info(a ...interface{}) {
	defaultLogger.log("INFO", a...)
}

// OK logs a message at OK level using the default logger.
func OK(a ...interface{}) {
	defaultLogger.log("OK", a...)
}

// Error logs a message at ERROR level using the default logger.
func Error(a ...interface{}) {
	defaultLogger.log("ERROR", a...)
}

// Fatal logs a message at FATAL level using the default logger and exits the program.
func Fatal(a ...interface{}) {
	defaultLogger.log("FATAL", a...)
}

// Warn logs a message at WARN level using the default logger.
func Warn(a ...interface{}) {
	defaultLogger.log("WARN", a...)
}

// Debug logs a message at DEBUG level using the default logger.
func Debug(a ...interface{}) {
	defaultLogger.log("DEBUG", a...)
}

// SetLogFile sets the log file and format for the default logger.
func SetLogFile(filePath string, format FileFormat) error {
	return defaultLogger.SetLogFile(filePath, format)
}

// Close closes any open file handles in the global configuration.
// This should be called before program exit to ensure proper cleanup.
func Close() error {
	config := GetConfiguration()
	return config.Close()
}
