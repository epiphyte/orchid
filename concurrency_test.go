package orchid

import (
	"os"
	"sync"
	"testing"
	"time"
)

func TestConcurrentLogging(t *testing.T) {
	var logger Logger
	testFile := "test_concurrent.log"

	// Initialize logger
	err := logger.Init("concurrent-test")
	if err != nil {
		t.Fatalf("Failed to init logger: %v", err)
	}

	// Set up file logging
	err = logger.SetLogFile(testFile, FormatTXT)
	if err != nil {
		t.Fatalf("Failed to set log file: %v", err)
	}
	defer os.Remove(testFile)

	const numGoroutines = 100
	const logsPerGoroutine = 50

	var wg sync.WaitGroup

	// Start multiple goroutines writing logs concurrently
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

	// Test passed if no race conditions occurred
	t.Log("Concurrent logging completed successfully")
}

func TestConcurrentInit(t *testing.T) {
	var logger Logger
	const numOperations = 50
	var wg sync.WaitGroup

	// Test concurrent Init operations
	for i := 0; i < numOperations; i++ {
		wg.Add(1)

		// Goroutine doing Init
		go func(id int) {
			defer wg.Done()
			err := logger.Init("test")
			if err != nil {
				t.Errorf("Init failed: %v", err)
			}
			time.Sleep(time.Millisecond) // Small delay to allow some logging
		}(i)
	}

	wg.Wait()
	t.Log("Concurrent Init completed successfully")
}

func TestGlobalLoggerConcurrency(t *testing.T) {
	testFile := "test_global_concurrent.log"

	// Initialize global logger
	Init("global-concurrent-test")
	err := SetLogFile(testFile, FormatJSON)
	if err != nil {
		t.Fatalf("Failed to set log file for global logger: %v", err)
	}
	defer os.Remove(testFile)

	const numGoroutines = 50
	const logsPerGoroutine = 20

	var wg sync.WaitGroup

	// Test all log levels concurrently using global functions
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
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
					// Test re-initialization during logging
					if j == 10 {
						Init("reinit-test")
						SetLogFile(testFile, FormatTXT)
					}
				}
			}
		}(i)
	}

	wg.Wait()
	t.Log("Global logger concurrency test completed successfully")
}

func TestConcurrentFileOperations(t *testing.T) {
	const numLoggers = 10
	const numLogs = 20

	var loggers [numLoggers]Logger
	var wg sync.WaitGroup

	// Start multiple loggers writing to different files concurrently
	for i := 0; i < numLoggers; i++ {
		wg.Add(1)
		go func(loggerID int) {
			defer wg.Done()

			testFile := "test_concurrent_" + string(rune('0'+loggerID)) + ".log"
			err := loggers[loggerID].Init("concurrent-logger")
			if err != nil {
				t.Errorf("Failed to init logger %d: %v", loggerID, err)
				return
			}
			err = loggers[loggerID].SetLogFile(testFile, FormatJSON)
			if err != nil {
				t.Errorf("Failed to set log file for logger %d: %v", loggerID, err)
				return
			}
			defer os.Remove(testFile)

			// Write logs
			for j := 0; j < numLogs; j++ {
				loggers[loggerID].Info("Logger", loggerID, "message", j)
				loggers[loggerID].Error("Logger", loggerID, "error", j)
			}
		}(i)
	}

	wg.Wait()
	t.Log("Concurrent file operations completed successfully")
}
