package orchid

import (
	"os"
	"path/filepath"
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
		{"leading whitespace", " test", false, ""},  // Should trim and succeed
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
			} else if err != nil {
				t.Errorf("Expected no error for module name '%s', got: %v", tc.moduleName, err)
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
	// Some cases use bare file names, so run inside a temporary directory
	// to keep any created files out of the repository.
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}
	t.Cleanup(func() {
		GetConfiguration().Reset()
		os.Chdir(origDir)
	})

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
			defer GetConfiguration().Reset()

			err := SetLogFile(tc.filePath, tc.format)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for filePath '%s' and format %d, but got nil", tc.filePath, tc.format)
				} else if !strings.Contains(err.Error(), tc.errorSubstr) {
					t.Errorf("Expected error to contain '%s', got: %v", tc.errorSubstr, err)
				}
			} else if err != nil {
				t.Errorf("Expected no error for filePath '%s' and format %d, got: %v", tc.filePath, tc.format, err)
			}
		})
	}
}

func TestFileWriteErrorHandling(t *testing.T) {
	resetConfigOnCleanup(t)
	GetConfiguration().Reset()

	var logger Logger
	if err := logger.Init("file-error-test"); err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}

	invalidPath := filepath.Join(t.TempDir(), "nonexistent", "directory", "test.log")
	if err := SetLogFile(invalidPath, FormatTXT); err == nil {
		t.Fatal("Expected error when setting invalid file path, but got nil")
	}
	if got := GetConfiguration().GetDefaultFile(); got != "" {
		t.Errorf("Failed SetLogFile must not leave a file configured, got %q", got)
	}
}

func TestFileWriteErrorRecovery(t *testing.T) {
	// Logging continues on the console even if file writing fails.
	console := captureLogOutput(t)
	errs := captureFileErrors(t)
	path := tempLogPath(t, "error_recovery.log")

	if err := SetLogFile(path, FormatTXT); err != nil {
		t.Fatalf("Failed to set log file: %v", err)
	}
	var logger Logger
	if err := logger.Init("error-recovery-test"); err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}
	logger.Info("This should work normally")

	// Close the handle behind the configuration's back so writes fail.
	GetConfiguration().getLogFile().Close()

	logger.Info("still logs to console")
	logger.Error("and so does this")

	out := console.String()
	if !strings.Contains(out, "still logs to console") || !strings.Contains(out, "and so does this") {
		t.Errorf("Console output missing messages after file failure: %q", out)
	}
	if !strings.Contains(errs.String(), "ORCHID FILE ERROR") {
		t.Errorf("Expected a file error report, got: %q", errs.String())
	}
}

func TestFileErrorReportedOncePerEpisode(t *testing.T) {
	silenceLogOutput(t)
	errs := captureFileErrors(t)
	path := tempLogPath(t, "report_once.log")

	if err := SetLogFile(path, FormatTXT); err != nil {
		t.Fatalf("Failed to set log file: %v", err)
	}
	var logger Logger
	if err := logger.Init("report-once"); err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}

	GetConfiguration().getLogFile().Close()
	for i := 0; i < 10; i++ {
		logger.Info("failing write", i)
	}
	if n := strings.Count(errs.String(), "ORCHID FILE ERROR"); n != 1 {
		t.Errorf("Expected exactly 1 error report for 10 failed writes, got %d:\n%s", n, errs.String())
	}

	// Reconfiguring starts a new episode.
	if err := SetLogFile(path, FormatTXT); err != nil {
		t.Fatalf("Failed to reset log file: %v", err)
	}
	logger.Info("works again")
	GetConfiguration().getLogFile().Close()
	logger.Info("fails again")
	if n := strings.Count(errs.String(), "ORCHID FILE ERROR"); n != 2 {
		t.Errorf("Expected a second report after reconfiguration, got %d:\n%s", n, errs.String())
	}
}

func TestInvalidFormatHandling(t *testing.T) {
	silenceLogOutput(t)
	errs := captureFileErrors(t)
	path := tempLogPath(t, "invalid_format.log")

	var logger Logger
	if err := logger.Init("format-test"); err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}

	// SetDefaultFormat performs no validation, so an out-of-range value
	// reaches the writer.
	config := GetConfiguration()
	if err := config.SetDefaultFile(path); err != nil {
		t.Fatalf("Failed to set file: %v", err)
	}
	config.SetDefaultFormat(FileFormat(99))

	logger.Info("Testing invalid format handling")
	logger.Info("second attempt")

	if !strings.Contains(errs.String(), "unsupported log format: 99") {
		t.Errorf("Expected unsupported format error, got: %q", errs.String())
	}
	if n := strings.Count(errs.String(), "ORCHID FILE ERROR"); n != 1 {
		t.Errorf("Expected 1 report, got %d", n)
	}
	lines := closeAndReadLines(t, path)
	if len(lines) != 0 {
		t.Errorf("Expected nothing written with an invalid format, got: %v", lines)
	}
}
