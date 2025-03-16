package log

func (l *DefaultLogger) Info(format string, a ...any) {
	LogInfo(l.ComponentName, format, a...)
}

func (l *DefaultLogger) Warning(format string, a ...any) {
	LogWarning(l.ComponentName, format, a...)
}

func (l *DefaultLogger) Error(format string, a ...any) {
	LogError(l.ComponentName, format, a...)
}

func (l *DefaultLogger) Success(format string, a ...any) {
	LogSuccess(l.ComponentName, format, a...)
}
