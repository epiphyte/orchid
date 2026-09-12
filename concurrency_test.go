package orchid

import (
	"encoding/json"
	"sync"
	"testing"
)

func TestConcurrentLogging(t *testing.T) {
	silenceLogOutput(t)
	path := tempLogPath(t, "concurrent.log")

	var logger Logger
	if err := logger.Init("concurrent-test"); err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}
	if err := SetLogFile(path, FormatTXT); err != nil {
		t.Fatalf("Failed to set log file: %v", err)
	}

	const numGoroutines = 100
	const logsPerGoroutine = 50
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < logsPerGoroutine; j++ {
				logger.Info("Goroutine", id, "log", j)
				logger.Error("Goroutine", id, "error", j)
				logger.Debug("Goroutine", id, "debug", j)
			}
		}(i)
	}
	wg.Wait()

	lines := closeAndReadLines(t, path)
	want := numGoroutines * logsPerGoroutine * 3
	if len(lines) != want {
		t.Errorf("Expected %d lines in log file, got %d", want, len(lines))
	}
}

func TestConcurrentInitAndLog(t *testing.T) {
	silenceLogOutput(t)

	var logger Logger
	if err := logger.Init("initial"); err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}

	const numOperations = 50
	var wg sync.WaitGroup

	// Init and log calls interleave on the same instance.
	for i := 0; i < numOperations; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			if err := logger.Init("test"); err != nil {
				t.Errorf("Init failed: %v", err)
			}
		}()
		go func(id int) {
			defer wg.Done()
			logger.Info("logging during init", id)
		}(i)
	}
	wg.Wait()
}

func TestGlobalLoggerConcurrency(t *testing.T) {
	silenceLogOutput(t)
	path := tempLogPath(t, "global_concurrent.log")

	if err := Init("global-concurrent-test"); err != nil {
		t.Fatalf("Failed to init global logger: %v", err)
	}
	if err := SetLogFile(path, FormatJSON); err != nil {
		t.Fatalf("Failed to set log file: %v", err)
	}

	const numGoroutines = 50
	const logsPerGoroutine = 20
	var wg sync.WaitGroup
	var logged int64
	var loggedMu sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			count := 0
			for j := 0; j < logsPerGoroutine; j++ {
				switch j % 6 {
				case 0:
					Info("Global goroutine", id, "info", j)
				case 1:
					OK("Global goroutine", id, "ok", j)
				case 2:
					Warn("Global goroutine", id, "warn", j)
				case 3:
					Error("Global goroutine", id, "error", j)
				case 4:
					Debug("Global goroutine", id, "debug", j)
				case 5:
					// Re-initialize and re-open the same file mid-run.
					if j == 11 {
						Init("reinit-test")
						SetLogFile(path, FormatTXT)
					}
					continue
				}
				count++
			}
			loggedMu.Lock()
			logged += int64(count)
			loggedMu.Unlock()
		}(i)
	}
	wg.Wait()

	lines := closeAndReadLines(t, path)
	if int64(len(lines)) != logged {
		t.Errorf("Expected %d lines in log file, got %d", logged, len(lines))
	}
}

func TestConcurrentMultipleLoggers(t *testing.T) {
	silenceLogOutput(t)
	path := tempLogPath(t, "concurrent_shared.log")

	const numLoggers = 10
	const numLogs = 20

	var loggers [numLoggers]Logger
	var wg sync.WaitGroup

	if err := SetLogFile(path, FormatJSON); err != nil {
		t.Fatalf("Failed to set global log file: %v", err)
	}

	for i := 0; i < numLoggers; i++ {
		wg.Add(1)
		go func(loggerID int) {
			defer wg.Done()

			moduleName := "logger-" + string(rune('0'+loggerID))
			if err := loggers[loggerID].Init(moduleName); err != nil {
				t.Errorf("Failed to init logger %d: %v", loggerID, err)
				return
			}
			for j := 0; j < numLogs; j++ {
				loggers[loggerID].Info("Logger", loggerID, "message", j)
				loggers[loggerID].Error("Logger", loggerID, "error", j)
			}
		}(i)
	}
	wg.Wait()

	lines := closeAndReadLines(t, path)
	want := numLoggers * numLogs * 2
	if len(lines) != want {
		t.Fatalf("Expected %d lines in log file, got %d", want, len(lines))
	}

	perModule := map[string]int{}
	for _, line := range lines {
		var entry struct{ Module, Severity string }
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("Interleaved write produced invalid JSON: %v\n%s", err, line)
		}
		perModule[entry.Module]++
	}
	if len(perModule) != numLoggers {
		t.Errorf("Expected %d distinct modules, got %d: %v", numLoggers, len(perModule), perModule)
	}
	for module, n := range perModule {
		if n != numLogs*2 {
			t.Errorf("Module %s wrote %d lines, expected %d", module, n, numLogs*2)
		}
	}
}
