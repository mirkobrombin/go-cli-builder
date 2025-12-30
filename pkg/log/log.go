package log

import (
	"fmt"
	"os"
)

// LogLevel represents the severity of a log message.
type LogLevel int

const (
	LogLevelInfo LogLevel = iota
	LogLevelSuccess
	LogLevelWarning
	LogLevelError
)

// Logger defines the interface for logging messages with different severity levels.
type Logger interface {
	Info(format string, a ...any)
	Success(format string, a ...any)
	Warning(format string, a ...any)
	Error(format string, a ...any)
}

type defaultLogger struct{}

// New creates a new instance of the default Logger.
//
// Example:
//
//	logger := log.New()
//	logger.Info("Starting application...")
func New() Logger {
	return &defaultLogger{}
}

// Info logs an informational message.
//
// Example:
//
//	logger.Info("System status: %s", "OK")
func (l *defaultLogger) Info(format string, a ...any) {
	logMessage(LogLevelInfo, "", format, a...)
}

// Success logs a success message.
//
// Example:
//
//	logger.Success("Operation completed successfully")
func (l *defaultLogger) Success(format string, a ...any) {
	logMessage(LogLevelSuccess, "", format, a...)
}

// Warning logs a warning message.
//
// Example:
//
//	logger.Warning("Disk space is running low")
func (l *defaultLogger) Warning(format string, a ...any) {
	logMessage(LogLevelWarning, "", format, a...)
}

// Error logs an error message.
//
// Example:
//
//	logger.Error("Failed to connect to database: %v", err)
func (l *defaultLogger) Error(format string, a ...any) {
	logMessage(LogLevelError, "", format, a...)
}

// logMessage prints a formatted message to standard output based on log level and debug mode.
func logMessage(level LogLevel, component string, format string, a ...any) {
	color := getLogLevelColor(level)
	coloredLevel := colorize(getLogLevelSymbol(level), color)

	coloredComponent := ""
	if os.Getenv("DEBUG") == "1" && component != "" {
		coloredComponent = colorize(component, "blue") + " "
	}

	fmt.Printf("%s %s%s\n", coloredLevel, coloredComponent, fmt.Sprintf(format, a...))
}

// getLogLevelColor returns the ANSI color code name associated with a log level.
func getLogLevelColor(level LogLevel) string {
	switch level {
	case LogLevelInfo:
		return "blue"
	case LogLevelWarning:
		return "yellow"
	case LogLevelError:
		return "red"
	case LogLevelSuccess:
		return "green"
	default:
		return "reset"
	}
}

// getLogLevelSymbol returns the unicode symbol associated with a log level.
func getLogLevelSymbol(level LogLevel) string {
	switch level {
	case LogLevelInfo:
		return "ℹ"
	case LogLevelWarning:
		return "⚠️"
	case LogLevelError:
		return "ø"
	case LogLevelSuccess:
		return "✓"
	default:
		return ""
	}
}

// colorize applies the specified ANSI color to the text.
func colorize(text string, color string) string {
	colors := map[string]string{
		"red":    "\033[31m",
		"green":  "\033[32m",
		"yellow": "\033[33m",
		"blue":   "\033[34m",
		"reset":  "\033[0m",
	}
	return colors[color] + text + colors["reset"]
}
