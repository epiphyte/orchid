package orchid

import (
	"os"
	"testing"
)

func TestResourceCleanup(t *testing.T) {
	// Test file operations with logger
	var logger Logger
	testFile := "test_resource_cleanup.log"

	// Initialize logger
	err := logger.Init("test")
	if err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}

	// Set up global file logging
	err = SetLogFile(testFile, FormatTXT)
	if err != nil {
		t.Fatalf("Failed to set log file: %v", err)
	}

	// Write a log message to ensure file is created
	logger.Info("Test message")

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Log file was not created")
	}

	// Clean up test file
	os.Remove(testFile)
}

func TestReInitialization(t *testing.T) {
	var logger Logger
	testFile1 := "test_reinit1.log"
	testFile2 := "test_reinit2.log"

	// First initialization
	err := logger.Init("test1")
	if err != nil {
		t.Fatalf("Failed to init logger first time: %v", err)
	}

	// Set first global file
	err = SetLogFile(testFile1, FormatTXT)
	if err != nil {
		t.Fatalf("Failed to set first log file: %v", err)
	}

	// Write to first file
	logger.Info("Message to file 1")

	// Change global file (affects all loggers)
	err = SetLogFile(testFile2, FormatJSON)
	if err != nil {
		t.Fatalf("Failed to set second log file: %v", err)
	}

	// Re-initialize with different module name
	err = logger.Init("test2")
	if err != nil {
		t.Fatalf("Failed to re-init logger: %v", err)
	}

	// Write to second file (uses current global config)
	logger.Info("Message to file 2")

	// Clean up
	os.Remove(testFile1)
	os.Remove(testFile2)
}

func TestGlobalLoggerFileOperations(t *testing.T) {
	testFile := "test_global_file.log"

	// Initialize global logger
	Init("global-test")

	// Set up file logging
	err := SetLogFile(testFile, FormatTXT)
	if err != nil {
		t.Fatalf("Failed to set log file for global logger: %v", err)
	}

	// Log a message
	Info("Global test message")

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Global log file was not created")
	}

	// Test proper cleanup
	err = Close()
	if err != nil {
		t.Errorf("Error closing global logger: %v", err)
	}

	// Clean up test file
	os.Remove(testFile)
}

func TestConfigurationClose(t *testing.T) {
	config := GetConfiguration()
	testFile := "test_config_close.log"

	// Set a file
	err := config.SetDefaultFile(testFile)
	if err != nil {
		t.Fatalf("Failed to set default file: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Log file was not created by configuration")
	}

	// Close should clean up the file handle
	err = config.Close()
	if err != nil {
		t.Errorf("Error closing configuration: %v", err)
	}

	// Verify configuration was reset
	if config.GetDefaultFile() != "" {
		t.Error("Default file should be empty after Close()")
	}

	// Clean up test file
	os.Remove(testFile)
}

func TestMultipleFileSetOperations(t *testing.T) {
	config := GetConfiguration()
	config.Reset() // Start clean

	testFile1 := "test_multiple_1.log"
	testFile2 := "test_multiple_2.log"

	// Set first file
	err := config.SetDefaultFile(testFile1)
	if err != nil {
		t.Fatalf("Failed to set first file: %v", err)
	}

	// Set second file (should close first)
	err = config.SetDefaultFile(testFile2)
	if err != nil {
		t.Fatalf("Failed to set second file: %v", err)
	}

	// Verify second file is active
	if config.GetDefaultFile() != testFile2 {
		t.Errorf("Expected %s, got %s", testFile2, config.GetDefaultFile())
	}

	// Close configuration
	config.Close()

	// Clean up test files
	os.Remove(testFile1)
	os.Remove(testFile2)
}
