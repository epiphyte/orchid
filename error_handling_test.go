package orchid

import (
	"os"
	"strings"
	"testing"
)

func TestLoggerInitValidation(t *testing.T) {
	var logger Logger

	testCases := []struct {
		name        string
		moduleName  string
		expectError bool
		errorSubstr string
	}{
		{"valid module name", "test-module", false, ""},
		{"empty string", "", true, "cannot be empty"},
		{"whitespace only", "   ", true, "cannot be empty"},
		{"leading whitespace", " test", false, ""}, // Should trim and succeed
		{"trailing whitespace", "test ", false, ""}, // Should trim and succeed
		{"too long", strings.Repeat("a", 60), true, "too long"},
		{"exactly 50 chars", strings.Repeat("a", 50), false, ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := logger.Init(tc.moduleName)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for module name '%s', but got nil", tc.moduleName)
				} else if !strings.Contains(err.Error(), tc.errorSubstr) {
					t.Errorf("Expected error to contain '%s', got: %v", tc.errorSubstr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for module name '%s', got: %v", tc.moduleName, err)
				}
			}
		})
	}
}

func TestGlobalInitValidation(t *testing.T) {
	testCases := []struct {
		name        string
		moduleName  string
		expectError bool
	}{
		{"valid module", "global-test", false},
		{"empty module", "", true},
		{"whitespace only", "   ", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := Init(tc.moduleName)

			if tc.expectError && err == nil {
				t.Errorf("Expected error for module name '%s', but got nil", tc.moduleName)
			} else if !tc.expectError && err != nil {
				t.Errorf("Expected no error for module name '%s', got: %v", tc.moduleName, err)
			}
		})
	}
}

func TestSetLogFileValidation(t *testing.T) {
	testCases := []struct {
		name        string
		filePath    string
		format      FileFormat
		expectError bool
		errorSubstr string
	}{
		{"valid file and format", "test.log", FormatTXT, false, ""},
		{"empty file path", "", FormatTXT, false, ""}, // Empty path disables file logging
		{"invalid format too low", "test.log", FileFormat(-1), true, "invalid log format"},
		{"invalid format too high", "test.log", FileFormat(99), true, "invalid log format"},
		{"file path with whitespace", " test.log ", FormatTXT, true, "leading or trailing whitespace"},
		{"file path with null byte", "test\x00.log", FormatTXT, true, "null bytes"},
		{"file path too long", strings.Repeat("a", 270), FormatTXT, true, "too long"},
		{"filename 255 chars", strings.Repeat("a", 255), FormatTXT, false, ""},
		{"filename 256 chars", strings.Repeat("a", 256), FormatTXT, true, "too long"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clean up any previous state
			defer func() {
				config := GetConfiguration()
				config.Reset()
			}()

			err := SetLogFile(tc.filePath, tc.format)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for filePath '%s' and format %d, but got nil", tc.filePath, tc.format)
				} else if !strings.Contains(err.Error(), tc.errorSubstr) {
					t.Errorf("Expected error to contain '%s', got: %v", tc.errorSubstr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for filePath '%s' and format %d, got: %v", tc.filePath, tc.format, err)
				}
			}
		})
	}
}

func TestFileWriteErrorHandling(t *testing.T) {
	// Create a test logger
	var logger Logger
	err := logger.Init("file-error-test")
	if err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}

	// Try to set up logging to a directory that doesn't exist
	invalidPath := "/nonexistent/directory/test.log"
	err = SetLogFile(invalidPath, FormatTXT)

	// The SetLogFile should fail when trying to create the file
	if err == nil {
		t.Errorf("Expected error when setting invalid file path, but got nil")
		// Clean up if somehow it worked
		config := GetConfiguration()
		config.Reset()
		os.Remove(invalidPath)
	}
}

func TestFileWriteErrorRecovery(t *testing.T) {
	// This test verifies that logging continues to work even if file writing fails
	testFile := "test_error_recovery.log"

	// Set up file logging
	err := SetLogFile(testFile, FormatTXT)
	if err != nil {
		t.Fatalf("Failed to set log file: %v", err)
	}

	defer func() {
		config := GetConfiguration()
		config.Reset()
		os.Remove(testFile)
	}()

	// Create a logger
	var logger Logger
	err = logger.Init("error-recovery-test")
	if err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}

	// This should work normally and write to file
	logger.Info("This should work normally")

	// Now close/remove the file to simulate a file error
	config := GetConfiguration()
	logFile := config.getLogFile()
	if logFile != nil {
		logFile.Close() // This will cause subsequent writes to fail
	}

	// This should continue to log to console even though file writing fails
	// We can't easily test the stderr output, but we can verify it doesn't crash
	logger.Info("This should still log to console despite file error")
	logger.Error("Error logging should also continue")

	// Test should complete without panicking
	t.Log("Error recovery test completed successfully")
}

func TestInvalidFormatHandling(t *testing.T) {
	var logger Logger
	err := logger.Init("format-test")
	if err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}

	// Set up file with an unusual format by bypassing validation
	config := GetConfiguration()
	config.mu.Lock()
	config.defaultFile = "test_invalid_format.log"
	config.defaultFormat = FileFormat(99) // Invalid format

	// Try to open the file
	logFile, err := os.OpenFile("test_invalid_format.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		config.mu.Unlock()
		t.Fatalf("Failed to create test file: %v", err)
	}
	config.logFile = logFile
	config.mu.Unlock()

	defer func() {
		config.Reset()
		os.Remove("test_invalid_format.log")
	}()

	// This should handle the invalid format gracefully
	logger.Info("Testing invalid format handling")

	// Test should complete without panicking
	t.Log("Invalid format test completed successfully")
}