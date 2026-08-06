# Logging

V3 provides a logging interface exposed via `cli.Base`. By embedding `cli.Base` in your command struct, you get access to a `Logger` instance.

## Interface

The `Logger` follows this interface:

```go
type Logger interface {
    Info(format string, a ...any)
    Success(format string, a ...any)
    Warning(format string, a ...any)
    Error(format string, a ...any)
}
```

## Usage

```go
func (c *MyCommand) Run() error {
    c.Logger.Info("Starting process...")

    if err := doSomething(); err != nil {
        c.Logger.Error("Something went wrong: %v", err)
        return err
    }

    c.Logger.Success("Done!")
    return nil
}
```

## What the default logger does

The default logger is built for a person reading a terminal:

- `Info` and `Success` are written to **stdout**, `Warning` and `Error` to **stderr**, so that piping a command's output does not hide its errors.
- Each line is prefixed with a symbol for its severity, colored when the destination can show colors.
- Colors are decided per stream. Redirecting stdout to a file leaves the errors on the terminal colored, and redirecting stderr to a file writes no escape codes into it.
- `NO_COLOR` and `TERM=dumb` disable colors entirely.
- A message logged without arguments is printed verbatim, so a line containing a percent sign survives.

## Structured output

When the output is read by a machine rather than a person, swap the logger for
the structured one. It emits one JSON object per line through the go-foundation
logger:

```go
import "github.com/mirkobrombin/go-cli-builder/v3/pkg/log/fdn"

cli.Run(app, cli.WithLogger(fdn.NewStructured(os.Stdout)))
```

`Success` has no equivalent severity in a structured logger, so it is written as
an `info` entry carrying a `status: success` field. The distinction is kept
rather than lost.

## Using a logger the application already has

If your application already builds a go-foundation logger, with its own level
and sinks, hand it over instead of building a second one:

```go
fl := logger.New(logger.WithLevel(logger.DebugLevel), logger.WithSink(mySink))
cli.Run(app, cli.WithLogger(fdn.Wrap(fl)))
```

Anything satisfying the four method interface is accepted by `cli.WithLogger`,
so a logger from another library only needs a small adapter.

## Testing command output

`log.NewWriter` returns a logger that writes to the writers you give it, with no
colors, which is what you want in a test:

```go
var out, errOut bytes.Buffer
cmd := &MyCommand{}
cmd.Logger = log.NewWriter(&out, &errOut)
```

## A note on v2

Between v2.1.0 and v2.2.1 the logger was backed directly by the go-foundation
console sink, which serializes entries as JSON. The practical effect was that
every CLI built on those versions printed lines like
`{"level":"info","time":"...","msg":"Installing..."}` to its users, and
`Success` was indistinguishable from `Info`. V3 fixes both.
