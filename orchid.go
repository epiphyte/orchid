// Package orchid
//Copyright (c) 2022 Epiphyte LLC. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
// Author: Fernandez-Alcon, Jose
// e-mail: jose@epiphyte.io
package orchid

import (
	"fmt"
	"log"
	"time"
)

var module = "NO_NAME"

const (
	COLOR_RESET = "\033[0m"
	COLOR_INFO  = "\033[48;5;33m"
	COLOR_OK    = "\033[48;5;36m"
	COLOR_WARN  = "\033[48;5;3m"
	COLOR_ERROR = "\033[48;5;1m"
	COLOR_FATAL = "\033[48;5;1m"
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

func (l *logMessage) createLogMessage(severity string, a ...interface{}) {
	l.Time = time.Now()
	l.Text = fmt.Sprint(a...)
	l.Severity = severity
}

func (l *logMessage) printLogMessage() {
	metadata := fmt.Sprintf("%-20s %-6s", module, l.Severity)
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
	case "FATAL":
		color = COLOR_FATAL
		break
	case "DEBUG":
		color = COLOR_DEBUG
		break
	}
	if l.Severity == "FATAL" {
		log.Fatal(string(COLOR_RESET), string(color), metadata, string(COLOR_RESET), l.Text)
	} else {
		log.Println(string(COLOR_RESET), string(color), metadata, string(COLOR_RESET), l.Text)
	}
}

func Info(a ...interface{}) {
	var l logMessage
	l.createLogMessage("INFO", a...)
	l.printLogMessage()
}

func OK(a ...interface{}) {
	var l logMessage
	l.createLogMessage("OK", a...)
	l.printLogMessage()
}

func Error(a ...interface{}) {
	var l logMessage
	l.createLogMessage("ERROR", a...)
	l.printLogMessage()
}

func Fatal(a ...interface{}) {
	var l logMessage
	l.createLogMessage("FATAL", a...)
	l.printLogMessage()
}

func Warn(a ...interface{}) {
	var l logMessage
	l.createLogMessage("WARN", a...)
	l.printLogMessage()
}

func Debug(a ...interface{}) {
	var l logMessage
	l.createLogMessage("DEBUG", a...)
	l.printLogMessage()
}
