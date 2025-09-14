package main

import (
	log "github.com/epiphyte/orchid"
)

func main() {
	// Initialize the default logger with a module name
	log.Init("example-app")
	log.SetLogFile("app.log", log.FormatTXT) // Set default log file and format

	// Demonstrate different log levels with console output
	log.Info("Application starting up")
	log.OK("Database connection established")
	log.Warn("Configuration file not found, using defaults")
	log.Error("Failed to connect to external API")
	log.Debug("Processing user request with ID: 12345")

	log.SetLogFile("app.json", log.FormatJSON) // Set default log file and format

	// Example with multiple arguments
	log.Info("User", "john_doe", "logged in from IP", "192.168.1.100")

	// Create a custom logger instance for a specific module
	var dbLogger log.Logger
	dbLogger.Init("database")
	dbLogger.Info("Database query executed successfully")
	dbLogger.OK("Transaction committed")

	// Example with file logging in different formats
	var fileLogger log.Logger
	err := fileLogger.Init("file-logger")
	if err != nil {
		log.Error("Failed to initialize file logger:", err)
	} else {
		fileLogger.Info("This message will be written to app.log in text format")
	}

	// Example with JSON file logging
	var jsonLogger log.Logger
	err = jsonLogger.Init("json-logger")
	if err != nil {
		log.Error("Failed to initialize JSON logger:", err)
	} else {
		jsonLogger.Info("This message will be written to app.json in JSON format")
		jsonLogger.OK("JSON logging is working properly")
	}

	log.Info("Example completed successfully")
}
