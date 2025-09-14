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

func (l *Logger) Init(moduleName, filePath string, format FileFormat) error {
	l.module = moduleName
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

func (l *Logger) log(severity string, a ...interface{}) {
	msg := l.createLogMessage(severity, a...)
	l.printLogMessage(msg)
}

func (l *Logger) Info(a ...interface{}) {
	l.log("INFO", a...)
}

func (l *Logger) OK(a ...interface{}) {
	l.log("OK", a...)
}

func (l *Logger) Error(a ...interface{}) {
	l.log("ERROR", a...)
}

func (l *Logger) Fatal(a ...interface{}) {
	l.log("FATAL", a...)
}

func (l *Logger) Warn(a ...interface{}) {
	l.log("WARN", a...)
}

func (l *Logger) Debug(a ...interface{}) {
	l.log("DEBUG", a...)
}

var defaultLogger Logger

func Init(moduleName string) {
	defaultLogger.Init(moduleName, "", FormatTXT)
}

func InitWithFile(moduleName, filePath string, format FileFormat) error {
	return defaultLogger.Init(moduleName, filePath, format)
}

func Info(a ...interface{}) {
	defaultLogger.log("INFO", a...)
}

func OK(a ...interface{}) {
	defaultLogger.log("OK", a...)
}

func Error(a ...interface{}) {
	defaultLogger.log("ERROR", a...)
}

func Fatal(a ...interface{}) {
	defaultLogger.log("FATAL", a...)
}

func Warn(a ...interface{}) {
	defaultLogger.log("WARN", a...)
}

func Debug(a ...interface{}) {
	defaultLogger.log("DEBUG", a...)
}
