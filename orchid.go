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

//logMessage struct describes a log message
type logMessage struct {
	Severity string    //The severity of the message [INFO, DEBUG, SUCCESS, WARNING, ERROR, FATAL]
	Text     string    //The contents of the log
	Module   string    //The name of the module where the log was originated
	Time     time.Time // The time at which the log was created
}

//Init initializes de module. It sets a name of the module calling the logger to filter the logs
func Init(module_name string) {
	module = module_name
}

//createLogMessage is internal helper to fill in the logMessage struct
func (l *logMessage) createLogMessage(severity string, a ...interface{}) {
	l.Time = time.Now()
	l.Text = fmt.Sprint(a...)
	l.Severity = severity
}

//printLogMessage is the internal function that colorizes and pretty prints the log message
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

//Info prints a log message of the severity level Info which has blue background on the terminal
func Info(a ...interface{}) {
	var l logMessage
	l.createLogMessage("INFO", a...)
	l.printLogMessage()
}

//OK prints a log message of the severity level OK which has green background on the terminal
func OK(a ...interface{}) {
	var l logMessage
	l.createLogMessage("OK", a...)
	l.printLogMessage()
}

//Error prints a log message of the severity level Error which has red background on the terminal
func Error(a ...interface{}) {
	var l logMessage
	l.createLogMessage("ERROR", a...)
	l.printLogMessage()
}

//Fatal prints a log message of the severity level Fatal which has red background on the terminal and terminates the
//execution of the program
func Fatal(a ...interface{}) {
	var l logMessage
	l.createLogMessage("FATAL", a...)
	l.printLogMessage()
}

//Warn prints a log message of the severity level Warn which has yellow background on the terminal
func Warn(a ...interface{}) {
	var l logMessage
	l.createLogMessage("WARN", a...)
	l.printLogMessage()
}

//Debug prints a log message of the severity level Debug which has purple background on the terminal
func Debug(a ...interface{}) {
	var l logMessage
	l.createLogMessage("DEBUG", a...)
	l.printLogMessage()
}
