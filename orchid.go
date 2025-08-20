package orchid

import (
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

type logMessage struct {
	Severity string
	Text     string
	Module   string
	Time     time.Time
}

type Logger struct {
	module  string
	logFile *os.File
}

func (l *Logger) Init(module_name, filePath string) error {
	l.module = module_name
	if filePath != "" {
		var err error
		l.logFile, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file: %v", err)
		}
		log.SetOutput(l.logFile)
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
	message := fmt.Sprintf("%s %s %s %s %s", COLOR_RESET, color, metadata, COLOR_RESET, msg.Text)
	if l.logFile != nil {
		fmt.Fprintln(l.logFile, message)
	}
	if msg.Severity == "FATAL" {
		log.Fatal(message)
	} else {
		log.Println(message)
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
