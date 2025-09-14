package orchid

import (
	"bytes"
	"log"
	"os"
	"sync"
	"testing"
	"time"
)

func TestGlobalLoggerThreadSafety(t *testing.T) {
	// Capture log output for verification
	var buf bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(originalOutput)

	// Initialize the global logger
	Init("thread-safety-test")

	const numGoroutines = 100
	const logsPerGoroutine = 50
	var wg sync.WaitGroup

	// Test concurrent calls to all global logging functions
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
					// Test SetLogFile concurrency as well
					if j%10 == 5 {
						SetLogFile("", FormatTXT) // Disable file logging
					}
				}
			}
		}(i)
	}

	wg.Wait()
	t.Log("Global logger thread safety test completed successfully")
}

func TestGlobalLoggerInitRace(t *testing.T) {
	// Test concurrent Init calls with logging
	const numGoroutines = 50
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			// Concurrent Init calls
			Init("race-test")

			// Immediate logging after Init
			Info("Goroutine", goroutineID, "initialized")
			Error("Goroutine", goroutineID, "error test")
		}(i)
	}

	wg.Wait()
	t.Log("Global logger Init race test completed successfully")
}

func TestGlobalLoggerFileOperationRace(t *testing.T) {
	testFile1 := "test_race_1.log"
	testFile2 := "test_race_2.log"

	defer func() {
		os.Remove(testFile1)
		os.Remove(testFile2)
		Close() // Clean up
	}()

	Init("file-race-test")

	const numGoroutines = 30
	var wg sync.WaitGroup

	// Test concurrent file operations with logging
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			// Alternate between different file configurations
			if goroutineID%2 == 0 {
				SetLogFile(testFile1, FormatTXT)
				Info("Goroutine", goroutineID, "using file 1")
			} else {
				SetLogFile(testFile2, FormatJSON)
				Error("Goroutine", goroutineID, "using file 2")
			}

			// Some goroutines disable file logging
			if goroutineID%5 == 0 {
				SetLogFile("", FormatTXT)
				Debug("Goroutine", goroutineID, "disabled file logging")
			}
		}(i)
	}

	wg.Wait()
	t.Log("Global logger file operation race test completed successfully")
}

func TestConfigurationConcurrency(t *testing.T) {
	config := GetConfiguration()
	config.Reset() // Start clean

	const numGoroutines = 50
	var wg sync.WaitGroup

	// Test concurrent configuration changes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			// Concurrent configuration operations
			config.SetEnableColors(goroutineID%2 == 0)
			config.SetDefaultFormat(FileFormat(goroutineID % 2))

			// Read operations
			_ = config.GetEnableColors()
			_ = config.GetDefaultFormat()
			_ = config.GetDefaultFile()

			// File operations
			if goroutineID%3 == 0 {
				testFile := "test_config_concurrent.log"
				config.SetDefaultFile(testFile)
				defer os.Remove(testFile)
			}
		}(i)
	}

	wg.Wait()
	config.Reset() // Clean up
	t.Log("Configuration concurrency test completed successfully")
}

func TestMixedGlobalAndInstanceLoggers(t *testing.T) {
	// Test that global logger and instance loggers don't interfere
	Init("global-mixed-test")

	const numGoroutines = 40
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			if goroutineID%2 == 0 {
				// Use global logger
				Info("Global goroutine", goroutineID)
				Error("Global error", goroutineID)
			} else {
				// Use instance logger
				var logger Logger
				logger.Init("instance-test")
				logger.Info("Instance goroutine", goroutineID)
				logger.Error("Instance error", goroutineID)
			}
		}(i)
	}

	wg.Wait()
	t.Log("Mixed global and instance logger test completed successfully")
}

func TestRapidInitAndLog(t *testing.T) {
	// Test rapid initialization and immediate logging
	const numIterations = 100
	var wg sync.WaitGroup

	for i := 0; i < numIterations; i++ {
		wg.Add(1)
		go func(iteration int) {
			defer wg.Done()

			// Rapid Init followed by immediate logging
			Init("rapid-test")
			Info("Rapid test", iteration)

			// Small delay to vary timing
			time.Sleep(time.Microsecond)

			Error("Rapid error", iteration)
		}(i)
	}

	wg.Wait()
	t.Log("Rapid init and log test completed successfully")
}