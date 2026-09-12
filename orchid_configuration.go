package orchid

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

// Configuration holds global configuration settings for the orchid logger.
// This singleton manages default file path, format, and other global settings.
type Configuration struct {
	mu                sync.RWMutex // Protects configuration fields
	defaultFile       string       // Default file path for logging
	defaultFormat     FileFormat   // Default format for file logging
	enableColors      bool         // Enable/disable color output
	logFile           *os.File     // Shared log file instance
	errOut            io.Writer    // Destination for file-logging error reports
	fileErrorReported bool         // True once a file error has been reported for the current failure episode
}

var (
	configInstance *Configuration
	configOnce     sync.Once
)

// GetConfiguration returns the singleton configuration instance.
// This function is thread-safe and uses lazy initialization.
func GetConfiguration() *Configuration {
	configOnce.Do(func() {
		configInstance = &Configuration{
			defaultFile:   "",        // No default file - console only
			defaultFormat: FormatTXT, // Default to text format
			enableColors:  true,      // Colors enabled by default
			errOut:        os.Stderr,
		}
	})
	return configInstance
}

func (c *Configuration) getLogFile() *os.File {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.logFile
}

// setErrorOutput redirects file-logging error reports. Used by tests.
func (c *Configuration) setErrorOutput(w io.Writer) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.errOut = w
}

// SetDefaultFile sets the default file path for all loggers.
// Pass empty string to disable file logging.
// The new file is opened before the previous one is closed, so if the new
// file cannot be opened the previous configuration is left untouched and
// an error is returned.
func (c *Configuration) SetDefaultFile(filePath string) error {
	var newFile *os.File
	if filePath != "" {
		f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		newFile = f
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.logFile != nil {
		c.logFile.Close()
	}
	c.logFile = newFile
	c.defaultFile = filePath
	c.fileErrorReported = false

	return nil
}

// GetDefaultFile returns the current default file path.
func (c *Configuration) GetDefaultFile() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.defaultFile
}

// SetDefaultFormat sets the default format for file logging.
func (c *Configuration) SetDefaultFormat(format FileFormat) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.defaultFormat = format
}

// GetDefaultFormat returns the current default file format.
func (c *Configuration) GetDefaultFormat() FileFormat {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.defaultFormat
}

// SetEnableColors enables or disables color output for console logging.
func (c *Configuration) SetEnableColors(enable bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.enableColors = enable
}

// GetEnableColors returns whether colors are enabled for console output.
func (c *Configuration) GetEnableColors() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.enableColors
}

// write formats msg and appends it to the log file, if one is configured.
// The lock is held for the duration of the write so the file handle cannot
// be closed or replaced underneath an in-flight write.
//
// Write failures never propagate to the caller. The first failure of an
// episode is reported to the error output; subsequent failures are silent
// until a write succeeds or the file is reconfigured.
func (c *Configuration) write(msg logMessage) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.defaultFile == "" {
		return // No file configured - nothing to do
	}

	err := c.writeLocked(msg)
	if err == nil {
		c.fileErrorReported = false
		return
	}
	if c.fileErrorReported {
		return
	}
	c.fileErrorReported = true
	fmt.Fprintf(c.errOut, "ORCHID FILE ERROR: %v (further file errors suppressed until a write succeeds or the log file is reconfigured)\n", err)
}

// writeLocked performs the actual formatting and write. c.mu must be held.
func (c *Configuration) writeLocked(msg logMessage) error {
	if c.logFile == nil {
		return errors.New("log file configured but file handle is not available")
	}

	line, err := formatFileLine(msg, c.defaultFormat)
	if err != nil {
		return err
	}

	if _, err := fmt.Fprintln(c.logFile, line); err != nil {
		return fmt.Errorf("failed to write log to file: %w", err)
	}
	return nil
}

// Close closes any open file handles and cleans up resources.
// After calling Close, the configuration can still be used but file logging
// will be disabled until SetDefaultFile is called again.
func (c *Configuration) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.logFile != nil {
		err := c.logFile.Close()
		c.logFile = nil
		c.defaultFile = ""
		c.fileErrorReported = false
		if err != nil {
			return fmt.Errorf("failed to close log file: %w", err)
		}
	}
	return nil
}

// Reset resets all configuration values to their defaults and closes any open files.
// This is primarily useful for testing.
func (c *Configuration) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Close existing file handle if open
	if c.logFile != nil {
		c.logFile.Close()
		c.logFile = nil
	}

	c.defaultFile = ""
	c.defaultFormat = FormatTXT
	c.enableColors = true
	c.errOut = os.Stderr
	c.fileErrorReported = false
}
