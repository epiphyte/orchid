package orchid

import (
	"os"
	"testing"
)

func TestResourceCleanup(t *testing.T) {
	// Test Close method on logger with file
	var logger Logger
	testFile := "test_resource_cleanup.log"

	// Initialize with file
	err := logger.Init("test", testFile, FormatTXT)
	if err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}

	// Write a log message to ensure file is created
	logger.Info("Test message")

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}

	// Close the logger
	err = logger.Close()
	if err != nil {
		t.Errorf("Error closing logger: %v", err)
	}

	// Verify file handle is nil
	if logger.logFile != nil {
		t.Error("Log file handle should be nil after Close()")
	}

	// Calling Close again should be safe
	err = logger.Close()
	if err != nil {
		t.Errorf("Second Close() should not error: %v", err)
	}

	// Clean up test file
	os.Remove(testFile)
}

func TestReInitialization(t *testing.T) {
	var logger Logger
	testFile1 := "test_reinit1.log"
	testFile2 := "test_reinit2.log"

	// First initialization
	err := logger.Init("test1", testFile1, FormatTXT)
	if err != nil {
		t.Fatalf("Failed to init logger first time: %v", err)
	}

	// Write to first file
	logger.Info("Message to file 1")

	// Re-initialize with different file (should close first file)
	err = logger.Init("test2", testFile2, FormatJSON)
	if err != nil {
		t.Fatalf("Failed to re-init logger: %v", err)
	}

	// Write to second file
	logger.Info("Message to file 2")

	// Clean up
	logger.Close()
	os.Remove(testFile1)
	os.Remove(testFile2)
}

func TestGlobalLoggerClose(t *testing.T) {
	testFile := "test_global_close.log"

	// Initialize global logger with file
	err := InitWithFile("global-test", testFile, FormatTXT)
	if err != nil {
		t.Fatalf("Failed to init global logger: %v", err)
	}

	// Log a message
	Info("Global test message")

	// Close global logger
	err = Close()
	if err != nil {
		t.Errorf("Error closing global logger: %v", err)
	}

	// Clean up test file
	os.Remove(testFile)
}