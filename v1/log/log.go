package log

import (
	"fmt"
	"os"

	"github.com/mirkobrombin/go-cli-builder/v1/utils"
)

// logMessage prints a formatted log message.
//
// Parameters:
//   - level: The log level.
//   - component: The component name.
//   - format: The format string.
//   - a: The arguments to format.
func LogMessage(level LogLevel, component string, format string, a ...any) {
	color := getLogLevelColor(level)
	coloredLevel := utils.Colorize(getLogLevelSymbol(level), color)

	coloredComponent := ""
	if os.Getenv("DEBUG") == "1" {
		coloredComponent = utils.Colorize(component, "blue")
	}

	fmt.Printf("%s %s %s\n", coloredLevel, coloredComponent, fmt.Sprintf(format, a...))
}

// GetLogLevelColor returns the color for a given log level.
//
// Parameters:
//   - level: The log level.
//
// Returns:
//   - The color associated with the log level.
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

// GetLogLevelSymbol returns the symbol for a given log level.
//
// Parameters:
//   - level: The log level.
//
// Returns:
//   - The symbol associated with the log level.
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
