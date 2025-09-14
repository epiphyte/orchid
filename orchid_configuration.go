package orchid

import (
	"fmt"
	"os"
	"sync"
)

// Configuration holds global configuration settings for the orchid logger.
// This singleton manages default file path, format, and other global settings.
type Configuration struct {
	mu            sync.RWMutex // Protects configuration fields
	defaultFile   string       // Default file path for logging
	defaultFormat FileFormat   // Default format for file logging
	enableColors  bool         // Enable/disable color output
	logFile       *os.File     // Shared log file instance
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
			defaultFile:   "",         // No default file - console only
			defaultFormat: FormatTXT,  // Default to text format
			enableColors:  true,       // Colors enabled by default
		}
	})
	return configInstance
}

func (c *Configuration) getLogFile() *os.File {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.logFile
}

// SetDefaultFile sets the default file path for all new loggers.
// Pass empty string to disable file logging by default.
// If a file is already open, it will be closed before opening the new one.
func (c *Configuration) SetDefaultFile(filePath string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Close existing file handle if open
	if c.logFile != nil {
		c.logFile.Close()
		c.logFile = nil
	}

	c.defaultFile = filePath

	// If filePath is empty, disable file logging
	if filePath == "" {
		return nil
	}

	var err error
	c.logFile, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

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
		if err != nil {
			return fmt.Errorf("failed to close log file: %v", err)
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
}
