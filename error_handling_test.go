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
	resetConfigOnCleanup(t)
	t.Chdir(t.TempDir())

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
		{"filename 255 chars", strings.Repeat("a", 255), FormatTXT, false, ""},
		{"filename 256 chars", strings.Repeat("a", 256), FormatTXT, true, "component too long"},
		{"component 270 chars inside path", "sub/" + strings.Repeat("a", 270), FormatTXT, true, "component too long"},
		{"whole path over 4096", strings.Repeat("abcdefghij/", 400) + "x.log", FormatTXT, true, "path too long"},
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

func TestSetLogFileAcceptsLongPathWithShortComponents(t *testing.T) {
	// A path longer than 255 bytes is valid as long as no single component
	// exceeds the limit.
	silenceLogOutput(t)
	dir := t.TempDir()
	resetConfigOnCleanup(t)
	for len(dir) < 300 {
		dir = filepath.Join(dir, strings.Repeat("d", 40))
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	path := filepath.Join(dir, "deep.log")

	if err := SetLogFile(path, FormatTXT); err != nil {
		t.Fatalf("Expected long path with short components to be accepted, got: %v", err)
	}
	var logger Logger
	if err := logger.Init("deep"); err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}
	logger.Info("deep write")
	if lines := closeAndReadLines(t, path); len(lines) != 1 {
		t.Errorf("Expected 1 line, got %v", lines)
	}
}

func TestConfigurationSettersValidate(t *testing.T) {
	// The Configuration setters are public and must enforce the same rules
	// as SetLogFile rather than relying on callers.
	resetConfigOnCleanup(t)
	config := GetConfiguration()
	config.Reset()

	if err := config.SetDefaultFile(" spaced.log"); err == nil {
		t.Error("Expected SetDefaultFile to reject leading whitespace")
	}
	if err := config.SetDefaultFile("nul\x00.log"); err == nil {
		t.Error("Expected SetDefaultFile to reject null bytes")
	}
	if config.GetDefaultFile() != "" {
		t.Errorf("Rejected path must not be stored, got %q", config.GetDefaultFile())
	}

	if err := config.SetDefaultFormat(FileFormat(99)); err == nil {
		t.Error("Expected SetDefaultFormat to reject an undefined format")
	}
	if err := config.SetDefaultFormat(FileFormat(-1)); err == nil {
		t.Error("Expected SetDefaultFormat to reject a negative format")
	}
	if config.GetDefaultFormat() != FormatTXT {
		t.Errorf("Rejected format must not be stored, got %d", config.GetDefaultFormat())
	}
	if err := config.SetDefaultFormat(FormatJSON); err != nil {
		t.Errorf("Expected FormatJSON to be accepted, got: %v", err)
	}
	if config.GetDefaultFormat() != FormatJSON {
		t.Errorf("Expected FormatJSON to be stored, got %d", config.GetDefaultFormat())
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

	// Bypass validation to exercise the writer's defensive default branch.
	config := GetConfiguration()
	if err := config.SetDefaultFile(path); err != nil {
		t.Fatalf("Failed to set file: %v", err)
	}
	config.mu.Lock()
	config.defaultFormat = FileFormat(99)
	config.mu.Unlock()

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
