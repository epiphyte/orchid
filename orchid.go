package orchid

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

const (
	COLOR_RESET = "\033[0m"
	COLOR_INFO  = "\033[48;5;33m"
	COLOR_OK    = "\033[48;5;36m"
	COLOR_WARN  = "\033[48;5;3m"
	COLOR_ERROR = "\033[48;5;1m"
	COLOR_FATAL = "\033[48;5;1m"
	COLOR_DEBUG = "\033[48;5;5m"
)

type FileFormat int

const (
	FormatTXT FileFormat = iota
	FormatJSON
)

type logMessage struct {
	Severity string
	Text     string
	Module   string
	Time     time.Time
}

type Logger struct {
	module     string
	logFile    *os.File
	fileFormat FileFormat
}

func (l *Logger) Init(module_name, filePath string, format FileFormat) error {
	l.module = module_name
	l.fileFormat = format
	if filePath != "" {
		var err error
		l.logFile, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file: %v", err)
		}
	}
	return nil
}

func (l *Logger) createLogMessage(severity string, a ...interface{}) logMessage {
	return logMessage{
		Severity: severity,
		Text:     fmt.Sprint(a...),
		Module:   l.module,
		Time:     time.Now(),
	}
}

func (l *Logger) writeToFile(msg logMessage) {
	switch l.fileFormat {
	case FormatTXT:
		txtMessage := fmt.Sprintf("%s [%s] %s: %s", 
			msg.Time.Format("2006-01-02 15:04:05"), 
			msg.Severity, 
			msg.Module, 
			msg.Text)
		fmt.Fprintln(l.logFile, txtMessage)
	case FormatJSON:
		jsonData, err := json.Marshal(msg)
		if err != nil {
			fmt.Fprintf(l.logFile, "Error marshaling log message to JSON: %v\n", err)
			return
		}
		fmt.Fprintln(l.logFile, string(jsonData))
	}
}

func (l *Logger) printLogMessage(msg logMessage) {
	metadata := fmt.Sprintf("%-20s %-6s", msg.Module, msg.Severity)
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
	consoleMessage := fmt.Sprintf("%s %s %s %s %s", COLOR_RESET, color, metadata, COLOR_RESET, msg.Text)
	
	if l.logFile != nil {
		l.writeToFile(msg)
	}
	
	if msg.Severity == "FATAL" {
		log.Fatal(consoleMessage)
	} else {
		log.Println(consoleMessage)
	}
}

func (l *Logger) Info(a ...interface{}) {
	msg := l.createLogMessage("INFO", a...)
	l.printLogMessage(msg)
}

func (l *Logger) OK(a ...interface{}) {
	msg := l.createLogMessage("OK", a...)
	l.printLogMessage(msg)
}

func (l *Logger) Error(a ...interface{}) {
	msg := l.createLogMessage("ERROR", a...)
	l.printLogMessage(msg)
}

func (l *Logger) Fatal(a ...interface{}) {
	msg := l.createLogMessage("FATAL", a...)
	l.printLogMessage(msg)
}

func (l *Logger) Warn(a ...interface{}) {
	msg := l.createLogMessage("WARN", a...)
	l.printLogMessage(msg)
}

func (l *Logger) Debug(a ...interface{}) {
	msg := l.createLogMessage("DEBUG", a...)
	l.printLogMessage(msg)
}

var defaultLogger Logger

func Init(module_name string) {
	defaultLogger.Init(module_name, "", FormatTXT)
}

func InitWithFile(module_name, filePath string, format FileFormat) error {
	return defaultLogger.Init(module_name, filePath, format)
}

func Info(a ...interface{}) {
	defaultLogger.Info(a...)
}

func OK(a ...interface{}) {
	defaultLogger.OK(a...)
}

func Error(a ...interface{}) {
	defaultLogger.Error(a...)
}

func Fatal(a ...interface{}) {
	defaultLogger.Fatal(a...)
}

func Warn(a ...interface{}) {
	defaultLogger.Warn(a...)
}

func Debug(a ...interface{}) {
	defaultLogger.Debug(a...)
}
