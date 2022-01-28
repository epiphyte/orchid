package orchid

import (
	"fmt"
	"time"
)

var module = "NO_NAME"

const (
	COLOR_RESET = "\033[0m"
	COLOR_INFO  = "\033[48;5;33m"
	COLOR_OK    = "\033[48;5;36m"
	COLOR_WARN  = "\033[48;5;3m"
	COLOR_ERROR = "\033[48;5;1m"
	COLOR_DEBUG = "\033[48;5;5m"
)

//Describes the structure of a log message
type logMessage struct {
	Severity string    //The severity of the message [INFO, DEBUG, SUCCESS, WARNING, ERROR, FATAL]
	Text     string    //The contents of the log
	Module   string    //The name of the module where the log was originated
	Time     time.Time // The time at which the log was created
}

func Init(module_name string) {
	module = module_name
}

func (l *logMessage) createLogMessage(severity string, text string, a ...interface{}) {
	l.Time = time.Now()
	l.Text = fmt.Sprintf(text, a...)
	l.Severity = severity
}

func (l *logMessage) printLogMessage() {
	metadata := fmt.Sprintf("%s %-20s %-6s", l.Time.Format(time.RFC822), module, l.Severity)
	color := COLOR_INFO
	switch l.Severity {
	case "INFO":
		color = COLOR_INFO
		break
	case "OK":
		color = COLOR_OK
		break
	case "WARN":
		color = COLOR_WARN
		break
	case "ERROR":
		color = COLOR_ERROR
		break
	case "DEBUG":
		color = COLOR_DEBUG
		break
	}
	fmt.Println(string(COLOR_RESET), string(color), metadata, string(COLOR_RESET), l.Text)
}

func Info(message string, a ...interface{}) {
	var l logMessage
	l.createLogMessage("INFO", message, a...)
	l.printLogMessage()
}

func OK(message string, a ...interface{}) {
	var l logMessage
	l.createLogMessage("OK", message, a...)
	l.printLogMessage()
}

func Error(message string, a ...interface{}) {
	var l logMessage
	l.createLogMessage("ERROR", message, a...)
	l.printLogMessage()
}

func Warn(message string, a ...interface{}) {
	var l logMessage
	l.createLogMessage("WARN", message, a...)
	l.printLogMessage()
}

func Debug(message string, a ...interface{}) {
	var l logMessage
	l.createLogMessage("DEBUG", message, a...)
	l.printLogMessage()
}
