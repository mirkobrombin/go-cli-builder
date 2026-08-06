// Package fdn bridges the CLI logging contract and the structured logger of
// go-foundation.
//
// It is the only package of go-cli-builder that names a Foundation type in an
// exported signature. Importing it is a choice: a CLI that is happy with the
// readable console output never pulls it in, and a future major of Foundation
// cannot reach the rest of the library through it.
package fdn

import (
	"fmt"
	"io"

	"github.com/mirkobrombin/go-cli-builder/v3/pkg/log"
	flogger "github.com/mirkobrombin/go-foundation/v2/core/logger"
)

// statusField carries the CLI severity that Foundation does not model.
// Success is an Info entry for Foundation, so the distinction that matters to
// a command line tool is kept as a structured field instead of being lost.
const statusField = "status"

// wrapped adapts a Foundation logger to the CLI logging contract.
type wrapped struct {
	inner flogger.Logger
}

// Wrap adapts an existing Foundation logger so it can be handed to
// cli.WithLogger. Sinks, level and bound fields of the wrapped logger are kept.
//
// Example:
//
//	fl := logger.New(logger.WithLevel(logger.DebugLevel))
//	cli.Run(app, cli.WithLogger(fdn.Wrap(fl)))
func Wrap(inner flogger.Logger) log.Logger {
	if inner == nil {
		return log.New()
	}
	return &wrapped{inner: inner}
}

// NewStructured returns a logger that emits one JSON object per line on w,
// suitable for a service, a CI job, or anything that reads logs with a parser
// rather than with eyes.
//
// Example:
//
//	cli.Run(app, cli.WithLogger(fdn.NewStructured(os.Stdout)))
func NewStructured(w io.Writer) log.Logger {
	inner := flogger.New(
		flogger.WithoutDefaultSink(),
		flogger.WithSink(flogger.NewConsoleSink(w)),
	)
	return &wrapped{inner: inner}
}

// Unwrap returns the Foundation logger behind l, and false when l was not
// produced by this package. Use it to add a sink or to shut the logger down.
func Unwrap(l log.Logger) (flogger.Logger, bool) {
	w, ok := l.(*wrapped)
	if !ok {
		return nil, false
	}
	return w.inner, true
}

func (l *wrapped) Info(format string, a ...any) {
	l.inner.Info(render(format, a...), status(log.LogLevelInfo))
}

func (l *wrapped) Success(format string, a ...any) {
	l.inner.Info(render(format, a...), status(log.LogLevelSuccess))
}

func (l *wrapped) Warning(format string, a ...any) {
	l.inner.Warn(render(format, a...), status(log.LogLevelWarning))
}

func (l *wrapped) Error(format string, a ...any) {
	l.inner.Error(render(format, a...), status(log.LogLevelError))
}

// render formats the message only when arguments were supplied, so that a
// message containing a percent sign survives being logged on its own.
func render(format string, a ...any) string {
	if len(a) == 0 {
		return format
	}
	return fmt.Sprintf(format, a...)
}

func status(level log.LogLevel) flogger.Field {
	return flogger.Field{Key: statusField, Value: level.String()}
}
