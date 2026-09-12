package main

import (
	"fmt"
	"os"

	log "github.com/epiphyte/orchid"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	// Release the log file handle on exit.
	defer log.Close()

	// Initialize the default logger with a module name.
	if err := log.Init("example-app"); err != nil {
		return err
	}

	// All loggers share one global log file. Start with plain text.
	if err := log.SetLogFile("app.log", log.FormatTXT); err != nil {
		return err
	}

	log.Info("Application starting up")
	log.OK("Database connection established")
	log.Warn("Configuration file not found, using defaults")
	log.Error("Failed to connect to external API")
	log.Debug("Processing user request with ID: 12345")

	// Arguments are joined with fmt.Sprint, so add your own spacing.
	log.Info("User ", "john_doe", " logged in from IP ", "192.168.1.100")

	// A logger instance carries its own module name but writes to the same
	// global file as the default logger.
	var dbLogger log.Logger
	if err := dbLogger.Init("database"); err != nil {
		return err
	}
	dbLogger.Info("Database query executed successfully")
	dbLogger.OK("Transaction committed")

	// Switching the global file affects every logger from this point on.
	// Lines are written as one JSON object per line with keys
	// severity, text, module and time.
	if err := log.SetLogFile("app.json", log.FormatJSON); err != nil {
		return err
	}

	var apiLogger log.Logger
	if err := apiLogger.Init("api"); err != nil {
		return err
	}
	apiLogger.Info("This line goes to app.json as JSON")
	dbLogger.Info("So does this one, even though dbLogger was created earlier")

	// Colors are on by default only when stderr is a terminal and NO_COLOR
	// is unset. They can be forced either way.
	log.GetConfiguration().SetEnableColors(false)
	log.Info("Example completed successfully (printed without colors)")

	return nil
}
