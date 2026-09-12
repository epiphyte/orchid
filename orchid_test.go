// Package orchid
// Copyright (c) 2022 Epiphyte LLC. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
// Author: Fernandez-Alcon, Jose
// e-mail: jose@epiphyte.io
package orchid

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestINFO(t *testing.T) {
	buf := captureLogOutput(t)

	var logger Logger
	if err := logger.Init("TestFramework"); err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}

	logger.Info("Test message")
	output := buf.String()
	if !strings.Contains(output, "TestFramework") {
		t.Errorf("Expected TestFramework in output, got: %s", output)
	}
	if !strings.Contains(output, "INFO") {
		t.Errorf("Expected INFO in output, got: %s", output)
	}
}

func TestAllLogLevels(t *testing.T) {
	buf := captureLogOutput(t)

	var logger Logger
	if err := logger.Init("TestModule"); err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}

	testCases := []struct {
		method   func(...interface{})
		expected string
		color    string
	}{
		{logger.Info, "INFO", COLOR_INFO},
		{logger.OK, "OK", COLOR_OK},
		{logger.Warn, "WARN", COLOR_WARN},
		{logger.Error, "ERROR", COLOR_ERROR},
		{logger.Debug, "DEBUG", COLOR_DEBUG},
	}

	for _, tc := range testCases {
		buf.Reset()
		tc.method("test message")
		output := buf.String()
		if !strings.Contains(output, tc.expected) {
			t.Errorf("Expected %s in output, got: %s", tc.expected, output)
		}
		if !strings.Contains(output, "TestModule") {
			t.Errorf("Expected TestModule in output, got: %s", output)
		}
		if !strings.Contains(output, tc.color) {
			t.Errorf("Expected color code %q for %s in output, got: %q", tc.color, tc.expected, output)
		}
	}
}

func TestGlobalLogger(t *testing.T) {
	buf := captureLogOutput(t)

	if err := Init("GlobalTest"); err != nil {
		t.Fatalf("Failed to init global logger: %v", err)
	}

	Info("global info message")
	output := buf.String()
	if !strings.Contains(output, "INFO") || !strings.Contains(output, "GlobalTest") {
		t.Errorf("Expected INFO and GlobalTest in output, got: %s", output)
	}
}

func TestConsoleFormatWithoutColors(t *testing.T) {
	buf := captureLogOutput(t)
	resetConfigOnCleanup(t)
	GetConfiguration().SetEnableColors(false)

	var logger Logger
	if err := logger.Init("plain"); err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}
	logger.Warn("no colors here")

	output := buf.String()
	if strings.Contains(output, "\033[") {
		t.Errorf("Expected no ANSI escape codes with colors disabled, got: %q", output)
	}
	if !strings.Contains(output, "plain") || !strings.Contains(output, "WARN") || !strings.Contains(output, "no colors here") {
		t.Errorf("Expected module, severity and text in output, got: %q", output)
	}
}

func TestTextFileFormat(t *testing.T) {
	silenceLogOutput(t)
	path := tempLogPath(t, "text.log")

	if err := SetLogFile(path, FormatTXT); err != nil {
		t.Fatalf("SetLogFile failed: %v", err)
	}

	var logger Logger
	if err := logger.Init("txt-module"); err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}
	logger.Info("hello", " ", "world")
	logger.Error("boom")

	lines := closeAndReadLines(t, path)
	if len(lines) != 2 {
		t.Fatalf("Expected 2 lines, got %d: %v", len(lines), lines)
	}

	pattern := regexp.MustCompile(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2} \[(INFO|ERROR)\] txt-module: .+$`)
	for _, line := range lines {
		if !pattern.MatchString(line) {
			t.Errorf("Line does not match text format: %q", line)
		}
	}
	if !strings.HasSuffix(lines[0], "[INFO] txt-module: hello world") {
		t.Errorf("Unexpected first line: %q", lines[0])
	}
	if !strings.HasSuffix(lines[1], "[ERROR] txt-module: boom") {
		t.Errorf("Unexpected second line: %q", lines[1])
	}
}

func TestJSONFileFormat(t *testing.T) {
	silenceLogOutput(t)
	path := tempLogPath(t, "json.log")

	if err := SetLogFile(path, FormatJSON); err != nil {
		t.Fatalf("SetLogFile failed: %v", err)
	}

	var logger Logger
	if err := logger.Init("json-module"); err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}
	before := time.Now().Add(-time.Second)
	logger.Debug("structured ", 42)

	lines := closeAndReadLines(t, path)
	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d: %v", len(lines), lines)
	}

	var entry struct {
		Severity string
		Text     string
		Module   string
		Time     time.Time
	}
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("Line is not valid JSON: %v\n%s", err, lines[0])
	}
	if entry.Severity != "DEBUG" || entry.Module != "json-module" || entry.Text != "structured 42" {
		t.Errorf("Unexpected JSON fields: %+v", entry)
	}
	if entry.Time.Before(before) {
		t.Errorf("Timestamp %v is older than expected", entry.Time)
	}
}
