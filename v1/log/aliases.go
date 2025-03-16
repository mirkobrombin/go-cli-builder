package log

// Print a info log message.
func LogInfo(component string, format string, a ...any) {
	LogMessage(LogLevelInfo, component, format, a...)
}

// Print a warning log message.
func LogWarning(component string, format string, a ...any) {
	LogMessage(LogLevelWarning, component, format, a...)
}

// Print a error log message.
func LogError(component string, format string, a ...any) {
	LogMessage(LogLevelError, component, format, a...)
}

// Print a success log message.
func LogSuccess(component string, format string, a ...any) {
	LogMessage(LogLevelSuccess, component, format, a...)
}
