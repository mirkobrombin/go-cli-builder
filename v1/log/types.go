package log

// DefaultLogger implementa l'interfaccia Logger.
type DefaultLogger struct {
	ComponentName string
}

// LogLevel represents the severity of a log message.
type LogLevel string

const (
	// LogLevelInfo represents an informational log message.
	LogLevelInfo LogLevel = "info"

	// LogLevelWarning represents a warning log message.
	LogLevelWarning LogLevel = "warn"

	// LogLevelError represents an error log message.
	LogLevelError LogLevel = "err"

	// LogLevelSuccess represents a success log message.
	LogLevelSuccess LogLevel = "ok"
)
