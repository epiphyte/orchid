package orchid

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResourceCleanup(t *testing.T) {
	silenceLogOutput(t)
	path := tempLogPath(t, "resource_cleanup.log")

	var logger Logger
	if err := logger.Init("test"); err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}
	if err := SetLogFile(path, FormatTXT); err != nil {
		t.Fatalf("Failed to set log file: %v", err)
	}
	logger.Info("Test message")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("Log file was not created")
	}
	if err := Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
	if GetConfiguration().GetDefaultFile() != "" {
		t.Error("Default file should be empty after Close()")
	}
	if GetConfiguration().getLogFile() != nil {
		t.Error("File handle should be nil after Close()")
	}
	// Logging after Close must still work (console only).
	logger.Info("after close")
}

func TestReInitialization(t *testing.T) {
	silenceLogOutput(t)
	dir := t.TempDir()
	resetConfigOnCleanup(t)
	file1 := filepath.Join(dir, "reinit1.log")
	file2 := filepath.Join(dir, "reinit2.log")

	var logger Logger
	if err := logger.Init("test1"); err != nil {
		t.Fatalf("Failed to init logger first time: %v", err)
	}
	if err := SetLogFile(file1, FormatTXT); err != nil {
		t.Fatalf("Failed to set first log file: %v", err)
	}
	logger.Info("Message to file 1")

	if err := SetLogFile(file2, FormatJSON); err != nil {
		t.Fatalf("Failed to set second log file: %v", err)
	}
	if err := logger.Init("test2"); err != nil {
		t.Fatalf("Failed to re-init logger: %v", err)
	}
	logger.Info("Message to file 2")

	lines2 := closeAndReadLines(t, file2)
	if len(lines2) != 1 {
		t.Fatalf("Expected 1 line in file 2, got %d: %v", len(lines2), lines2)
	}
	var entry struct{ Module, Text string }
	if err := json.Unmarshal([]byte(lines2[0]), &entry); err != nil {
		t.Fatalf("File 2 line is not JSON: %v", err)
	}
	if entry.Module != "test2" || entry.Text != "Message to file 2" {
		t.Errorf("Unexpected file 2 entry: %+v", entry)
	}

	data1, err := os.ReadFile(file1)
	if err != nil {
		t.Fatalf("Failed to read file 1: %v", err)
	}
	if !strings.Contains(string(data1), "[INFO] test1: Message to file 1") {
		t.Errorf("Unexpected file 1 content: %q", data1)
	}
	if strings.Contains(string(data1), "file 2") {
		t.Errorf("File 1 received a message meant for file 2: %q", data1)
	}
}

func TestGlobalLoggerFileOperations(t *testing.T) {
	silenceLogOutput(t)
	path := tempLogPath(t, "global_file.log")

	if err := Init("global-test"); err != nil {
		t.Fatalf("Failed to init global logger: %v", err)
	}
	if err := SetLogFile(path, FormatTXT); err != nil {
		t.Fatalf("Failed to set log file for global logger: %v", err)
	}
	Info("Global test message")

	lines := closeAndReadLines(t, path)
	if len(lines) != 1 || !strings.HasSuffix(lines[0], "[INFO] global-test: Global test message") {
		t.Errorf("Unexpected file content: %v", lines)
	}
}

func TestConfigurationClose(t *testing.T) {
	path := tempLogPath(t, "config_close.log")
	config := GetConfiguration()

	if err := config.SetDefaultFile(path); err != nil {
		t.Fatalf("Failed to set default file: %v", err)
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("Log file was not created by configuration")
	}

	if err := config.Close(); err != nil {
		t.Errorf("Error closing configuration: %v", err)
	}
	if config.GetDefaultFile() != "" {
		t.Error("Default file should be empty after Close()")
	}
	// A second Close is a no-op.
	if err := config.Close(); err != nil {
		t.Errorf("Second Close should succeed, got: %v", err)
	}
}

func TestMultipleFileSetOperations(t *testing.T) {
	dir := t.TempDir()
	resetConfigOnCleanup(t)
	file1 := filepath.Join(dir, "multiple_1.log")
	file2 := filepath.Join(dir, "multiple_2.log")
	config := GetConfiguration()

	if err := config.SetDefaultFile(file1); err != nil {
		t.Fatalf("Failed to set first file: %v", err)
	}
	first := config.getLogFile()

	if err := config.SetDefaultFile(file2); err != nil {
		t.Fatalf("Failed to set second file: %v", err)
	}
	if config.GetDefaultFile() != file2 {
		t.Errorf("Expected %s, got %s", file2, config.GetDefaultFile())
	}
	// The first handle must have been closed when it was replaced.
	if _, err := first.Write([]byte("x")); err == nil {
		t.Error("Expected write to the replaced file handle to fail")
	}
}

func TestSetDefaultFileFailureKeepsPreviousFile(t *testing.T) {
	silenceLogOutput(t)
	errs := captureFileErrors(t)
	path := tempLogPath(t, "keep_previous.log")

	if err := SetLogFile(path, FormatTXT); err != nil {
		t.Fatalf("Failed to set log file: %v", err)
	}
	var logger Logger
	if err := logger.Init("keep"); err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}
	logger.Info("before failure")

	invalid := filepath.Join(t.TempDir(), "missing", "dir", "x.log")
	if err := SetLogFile(invalid, FormatJSON); err == nil {
		t.Fatal("Expected SetLogFile to fail for a path in a missing directory")
	}

	config := GetConfiguration()
	if config.GetDefaultFile() != path {
		t.Errorf("Expected default file to remain %s, got %q", path, config.GetDefaultFile())
	}
	if config.GetDefaultFormat() != FormatTXT {
		t.Errorf("Expected format to remain TXT, got %d", config.GetDefaultFormat())
	}

	logger.Info("after failure")

	lines := closeAndReadLines(t, path)
	if len(lines) != 2 {
		t.Errorf("Expected both messages in the original file, got %d lines: %v", len(lines), lines)
	}
	if errs.Len() != 0 {
		t.Errorf("Expected no file errors, got: %s", errs.String())
	}
}

func TestSetDefaultFileFailureFromCleanState(t *testing.T) {
	resetConfigOnCleanup(t)
	config := GetConfiguration()
	config.Reset()

	invalid := filepath.Join(t.TempDir(), "missing", "dir", "x.log")
	if err := config.SetDefaultFile(invalid); err == nil {
		t.Fatal("Expected SetDefaultFile to fail")
	}
	if config.GetDefaultFile() != "" {
		t.Errorf("Expected default file to stay empty after failed open, got %q", config.GetDefaultFile())
	}
	if config.getLogFile() != nil {
		t.Error("Expected no file handle after failed open")
	}
}
