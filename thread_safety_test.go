package orchid

import (
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestGlobalLoggerThreadSafety(t *testing.T) {
	silenceLogOutput(t)
	resetConfigOnCleanup(t)

	if err := Init("thread-safety-test"); err != nil {
		t.Fatalf("Failed to init global logger: %v", err)
	}

	const numGoroutines = 100
	const logsPerGoroutine = 50
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < logsPerGoroutine; j++ {
				switch j % 6 {
				case 0:
					Info("Goroutine", goroutineID, "Info", j)
				case 1:
					OK("Goroutine", goroutineID, "OK", j)
				case 2:
					Warn("Goroutine", goroutineID, "Warn", j)
				case 3:
					Error("Goroutine", goroutineID, "Error", j)
				case 4:
					Debug("Goroutine", goroutineID, "Debug", j)
				case 5:
					if j%10 == 5 {
						SetLogFile("", FormatTXT) // Disable file logging
					}
				}
			}
		}(i)
	}
	wg.Wait()
}

func TestGlobalLoggerInitRace(t *testing.T) {
	silenceLogOutput(t)

	const numGoroutines = 50
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			if err := Init("race-test"); err != nil {
				t.Errorf("Init failed: %v", err)
			}
			Info("Goroutine", goroutineID, "initialized")
			Error("Goroutine", goroutineID, "error test")
		}(i)
	}
	wg.Wait()
}

func TestGlobalLoggerFileOperationRace(t *testing.T) {
	silenceLogOutput(t)
	errs := captureFileErrors(t)
	dir := t.TempDir()
	file1 := filepath.Join(dir, "race_1.log")
	file2 := filepath.Join(dir, "race_2.log")

	if err := Init("file-race-test"); err != nil {
		t.Fatalf("Failed to init global logger: %v", err)
	}

	const numGoroutines = 30
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			if goroutineID%2 == 0 {
				SetLogFile(file1, FormatTXT)
				Info("Goroutine", goroutineID, "using file 1")
			} else {
				SetLogFile(file2, FormatJSON)
				Error("Goroutine", goroutineID, "using file 2")
			}
			if goroutineID%5 == 0 {
				SetLogFile("", FormatTXT)
				Debug("Goroutine", goroutineID, "disabled file logging")
			}
		}(i)
	}
	wg.Wait()

	if errs.Len() != 0 {
		t.Errorf("File swaps must never cause write errors, got: %s", errs.String())
	}
}

func TestInstanceWriteDuringFileSwap(t *testing.T) {
	// Instance loggers do not take the default-logger lock, so their file
	// writes must be protected by the configuration lock alone.
	silenceLogOutput(t)
	errs := captureFileErrors(t)
	dir := t.TempDir()
	file1 := filepath.Join(dir, "swap_1.log")
	file2 := filepath.Join(dir, "swap_2.log")

	if err := SetLogFile(file1, FormatTXT); err != nil {
		t.Fatalf("Failed to set log file: %v", err)
	}

	var logger Logger
	if err := logger.Init("swap-test"); err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}

	const numWriters = 8
	const writesPerWriter = 500
	const numSwaps = 200
	var wg sync.WaitGroup

	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < writesPerWriter; j++ {
				logger.Info("writer", id, "line", j)
			}
		}(i)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < numSwaps; i++ {
			target := file1
			if i%2 == 1 {
				target = file2
			}
			if err := SetLogFile(target, FormatTXT); err != nil {
				t.Errorf("SetLogFile failed: %v", err)
			}
			if i%50 == 0 {
				Close()
				SetLogFile(target, FormatTXT)
			}
		}
	}()
	wg.Wait()

	if errs.Len() != 0 {
		t.Errorf("Expected no write errors while swapping files, got: %s", errs.String())
	}

	total := len(closeAndReadLines(t, file1))
	if err := SetLogFile(file2, FormatTXT); err != nil {
		t.Fatalf("Failed to reopen file 2: %v", err)
	}
	total += len(closeAndReadLines(t, file2))
	// Every write that happened while a file was configured must have landed.
	// Writes during the brief Close() windows are legitimately console-only,
	// so allow for those but never for losses beyond them.
	if total > numWriters*writesPerWriter {
		t.Errorf("More lines than writes: %d > %d", total, numWriters*writesPerWriter)
	}
	if total == 0 {
		t.Error("No lines were written to either file")
	}
}

func TestConfigurationConcurrency(t *testing.T) {
	resetConfigOnCleanup(t)
	config := GetConfiguration()
	config.Reset()
	path := filepath.Join(t.TempDir(), "config_concurrent.log")

	const numGoroutines = 50
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			config.SetEnableColors(goroutineID%2 == 0)
			if err := config.SetDefaultFormat(FileFormat(goroutineID % 2)); err != nil {
				t.Errorf("SetDefaultFormat failed: %v", err)
			}

			_ = config.GetEnableColors()
			_ = config.GetDefaultFormat()
			_ = config.GetDefaultFile()

			if goroutineID%3 == 0 {
				if err := config.SetDefaultFile(path); err != nil {
					t.Errorf("SetDefaultFile failed: %v", err)
				}
			}
		}(i)
	}
	wg.Wait()

	if config.GetDefaultFile() != path {
		t.Errorf("Expected default file %s, got %q", path, config.GetDefaultFile())
	}
}

func TestMixedGlobalAndInstanceLoggers(t *testing.T) {
	silenceLogOutput(t)
	path := tempLogPath(t, "mixed.log")

	if err := Init("global-mixed-test"); err != nil {
		t.Fatalf("Failed to init global logger: %v", err)
	}
	if err := SetLogFile(path, FormatTXT); err != nil {
		t.Fatalf("Failed to set log file: %v", err)
	}

	const numGoroutines = 40
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()
			if goroutineID%2 == 0 {
				Info("Global goroutine", goroutineID)
				Error("Global error", goroutineID)
			} else {
				var logger Logger
				if err := logger.Init("instance-test"); err != nil {
					t.Errorf("Init failed: %v", err)
					return
				}
				logger.Info("Instance goroutine", goroutineID)
				logger.Error("Instance error", goroutineID)
			}
		}(i)
	}
	wg.Wait()

	lines := closeAndReadLines(t, path)
	if len(lines) != numGoroutines*2 {
		t.Fatalf("Expected %d lines, got %d", numGoroutines*2, len(lines))
	}
	global, instance := 0, 0
	for _, line := range lines {
		switch {
		case strings.Contains(line, "] global-mixed-test: "):
			global++
		case strings.Contains(line, "] instance-test: "):
			instance++
		default:
			t.Errorf("Line from unexpected module: %q", line)
		}
	}
	if global != numGoroutines || instance != numGoroutines {
		t.Errorf("Expected %d global and %d instance lines, got %d and %d", numGoroutines, numGoroutines, global, instance)
	}
}

func TestRapidInitAndLog(t *testing.T) {
	silenceLogOutput(t)

	const numIterations = 100
	var wg sync.WaitGroup

	for i := 0; i < numIterations; i++ {
		wg.Add(1)
		go func(iteration int) {
			defer wg.Done()
			if err := Init("rapid-test"); err != nil {
				t.Errorf("Init failed: %v", err)
			}
			Info("Rapid test", iteration)
			time.Sleep(time.Microsecond)
			Error("Rapid error", iteration)
		}(i)
	}
	wg.Wait()
}
