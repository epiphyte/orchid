// Package orchid
// Copyright (c) 2022 Epiphyte LLC. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
// Author: Fernandez-Alcon, Jose
// e-mail: jose@epiphyte.io
package orchid

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestINFO(t *testing.T) {
	var buf bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(originalOutput) // Reset to original

	var logger Logger

	err := logger.Init("TestFramework")
	if err != nil {
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
	var buf bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(originalOutput)

	var logger Logger
	err := logger.Init("TestModule")
	if err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}

	testCases := []struct {
		method   func(...interface{})
		expected string
	}{
		{logger.Info, "INFO"},
		{logger.OK, "OK"},
		{logger.Warn, "WARN"},
		{logger.Error, "ERROR"},
		{logger.Debug, "DEBUG"},
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
	}
}

func TestGlobalLogger(t *testing.T) {
	var buf bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(originalOutput)

	Init("GlobalTest")

	buf.Reset()
	Info("global info message")
	output := buf.String()
	if !strings.Contains(output, "INFO") || !strings.Contains(output, "GlobalTest") {
		t.Errorf("Expected INFO and GlobalTest in output, got: %s", output)
	}
}
