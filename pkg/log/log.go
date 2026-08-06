// Package log provides the logging contract used by every command, plus the
// two implementations a CLI normally needs: a human readable one for the
// terminal and a structured one for machines.
package log

import (
	"fmt"
	"io"
	"os"
)

// LogLevel represents the severity of a log message.
type LogLevel int

// The severities a CLI distinguishes. Success is deliberately separate from
// Info: for a command line tool the outcome of an operation is not the same
// kind of message as a progress note.
const (
	LogLevelInfo LogLevel = iota
	LogLevelSuccess
	LogLevelWarning
	LogLevelError
)

// Logger defines the interface for logging messages with different severity levels.
//
// Commands receive one through cli.Base. Supply your own with cli.WithLogger
// when the application already has a logging stack.
type Logger interface {
	Info(format string, a ...any)
	Success(format string, a ...any)
	Warning(format string, a ...any)
	Error(format string, a ...any)
}

// consoleLogger writes human readable lines: informational output on out,
// problems on errOut, so that piping stdout does not swallow errors.
//
// Colors are decided per writer, because stdout and stderr are redirected
// independently: "cmd > file" must still colorize the errors on the terminal,
// and "cmd 2> file" must not put escape codes in the file.
type consoleLogger struct {
	out         io.Writer
	errOut      io.Writer
	colorOut    bool
	colorErrOut bool
}

// New creates the default Logger: readable lines on the terminal, colors only
// when the terminal can show them.
//
// Example:
//
//	logger := log.New()
//	logger.Info("Starting application...")
func New() Logger {
	return &consoleLogger{
		out:         os.Stdout,
		errOut:      os.Stderr,
		colorOut:    colorEnabled(os.Stdout),
		colorErrOut: colorEnabled(os.Stderr),
	}
}

// NewWriter creates a human readable Logger writing to the given writers,
// without colors. Useful in tests and when the output is captured.
//
// Example:
//
//	var buf bytes.Buffer
//	logger := log.NewWriter(&buf, &buf)
func NewWriter(out, errOut io.Writer) Logger {
	if out == nil {
		out = io.Discard
	}
	if errOut == nil {
		errOut = out
	}
	return &consoleLogger{out: out, errOut: errOut}
}

func (l *consoleLogger) Info(format string, a ...any) {
	l.write(l.out, l.colorOut, LogLevelInfo, format, a...)
}

func (l *consoleLogger) Success(format string, a ...any) {
	l.write(l.out, l.colorOut, LogLevelSuccess, format, a...)
}

func (l *consoleLogger) Warning(format string, a ...any) {
	l.write(l.errOut, l.colorErrOut, LogLevelWarning, format, a...)
}

func (l *consoleLogger) Error(format string, a ...any) {
	l.write(l.errOut, l.colorErrOut, LogLevelError, format, a...)
}

func (l *consoleLogger) write(w io.Writer, color bool, level LogLevel, format string, a ...any) {
	msg := format
	if len(a) > 0 {
		msg = fmt.Sprintf(format, a...)
	}

	symbol := levelSymbol(level)
	if color {
		symbol = colorize(symbol, levelColor(level))
	}

	fmt.Fprintf(w, "%s %s\n", symbol, msg)
}

// colorEnabled reports whether ANSI colors should be emitted on w. Colors are
// off when the output is redirected to a file or a pipe, when NO_COLOR is set,
// and on terminals that declare themselves incapable.
//
// The check is the character device test rather than a real ioctl, to keep the
// library free of dependencies. The one case it gets wrong is a redirect to
// /dev/null, which is a character device too: colors are emitted there and
// then discarded, so nothing observable changes.
func colorEnabled(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}

	f, ok := w.(*os.File)
	if !ok {
		return false
	}

	info, err := f.Stat()
	if err != nil {
		return false
	}

	return info.Mode()&os.ModeCharDevice != 0
}

// levelColor returns the ANSI color name associated with a log level.
func levelColor(level LogLevel) string {
	switch level {
	case LogLevelInfo:
		return "blue"
	case LogLevelSuccess:
		return "green"
	case LogLevelWarning:
		return "yellow"
	case LogLevelError:
		return "red"
	default:
		return "reset"
	}
}

// levelSymbol returns the symbol printed in front of a message.
func levelSymbol(level LogLevel) string {
	switch level {
	case LogLevelInfo:
		return "ℹ"
	case LogLevelSuccess:
		return "✓"
	case LogLevelWarning:
		return "⚠"
	case LogLevelError:
		return "ø"
	default:
		return " "
	}
}

// String returns the textual name of the level, as used by structured output.
func (l LogLevel) String() string {
	switch l {
	case LogLevelInfo:
		return "info"
	case LogLevelSuccess:
		return "success"
	case LogLevelWarning:
		return "warning"
	case LogLevelError:
		return "error"
	default:
		return "unknown"
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
