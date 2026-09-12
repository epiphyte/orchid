package orchid

import (
	"bytes"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// resetConfigOnCleanup restores the global configuration (closing any open
// log file) when the test finishes, so state never leaks between tests.
func resetConfigOnCleanup(t *testing.T) {
	t.Helper()
	t.Cleanup(func() { GetConfiguration().Reset() })
}

// tempLogPath returns a path inside a per-test temporary directory and
// registers configuration reset. The reset is registered before the
// directory is created so it runs before the directory is removed.
func tempLogPath(t *testing.T, name string) string {
	t.Helper()
	resetConfigOnCleanup(t)
	return filepath.Join(t.TempDir(), name)
}

// captureLogOutput redirects the standard log package into a buffer for the
// duration of the test and returns it.
func captureLogOutput(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	orig := log.Writer()
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(orig) })
	return &buf
}

// silenceLogOutput discards console output for the duration of the test.
// Used by high-volume concurrency tests.
func silenceLogOutput(t *testing.T) {
	t.Helper()
	orig := log.Writer()
	log.SetOutput(io.Discard)
	t.Cleanup(func() { log.SetOutput(orig) })
}

// captureFileErrors redirects ORCHID FILE ERROR reports into a buffer.
func captureFileErrors(t *testing.T) *bytes.Buffer {
	t.Helper()
	resetConfigOnCleanup(t)
	var buf bytes.Buffer
	GetConfiguration().setErrorOutput(&buf)
	return &buf
}

// closeAndReadLines closes the global log file and returns the non-empty
// lines written to path.
func closeAndReadLines(t *testing.T, path string) []string {
	t.Helper()
	if err := GetConfiguration().Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read log file %s: %v", path, err)
	}
	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}
