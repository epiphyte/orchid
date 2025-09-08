package main

import (
	"github.com/epiphyte/orchid"
)

func main() {
	// Initialize the default logger with a module name
	orchid.Init("example-app")

	// Demonstrate different log levels with console output
	orchid.Info("Application starting up")
	orchid.OK("Database connection established")
	orchid.Warn("Configuration file not found, using defaults")
	orchid.Error("Failed to connect to external API")
	orchid.Debug("Processing user request with ID: 12345")

	// Example with multiple arguments
	orchid.Info("User", "john_doe", "logged in from IP", "192.168.1.100")

	// Create a custom logger instance for a specific module
	var dbLogger orchid.Logger
	dbLogger.Init("database", "", orchid.FormatTXT)
	dbLogger.Info("Database query executed successfully")
	dbLogger.OK("Transaction committed")

	// Example with file logging in different formats
	var fileLogger orchid.Logger
	err := fileLogger.Init("file-logger", "app.log", orchid.FormatTXT)
	if err != nil {
		orchid.Error("Failed to initialize file logger:", err)
	} else {
		fileLogger.Info("This message will be written to app.log in text format")
	}

	// Example with JSON file logging
	var jsonLogger orchid.Logger
	err = jsonLogger.Init("json-logger", "app.json", orchid.FormatJSON)
	if err != nil {
		orchid.Error("Failed to initialize JSON logger:", err)
	} else {
		jsonLogger.Info("This message will be written to app.json in JSON format")
		jsonLogger.OK("JSON logging is working properly")
	}

	orchid.Info("Example completed successfully")
}