# Logging

V2 provides a built-in logging interface exposed via `cli.Base`. By embedding `cli.Base` in your command struct, you get access to a `Logger` instance.

## Interface

The `Logger` follows this interface:

```go
type Logger interface {
    Info(format string, a ...interface{})
    Success(format string, a ...interface{})
    Warning(format string, a ...interface{})
    Error(format string, a ...interface{})
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

The logger automatically colorizes output (e.g., green for Success, red for Error) for better readability.