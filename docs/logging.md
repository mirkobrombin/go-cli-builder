# Logging

The Go CLI Builder library includes a basic logging system to help you track 
the execution of your CLI application and output messages to the user.

## Log Levels

The library defines the following log levels in the `log` package:

- `LogLevelInfo`: For informational messages.
- `LogLevelWarning`: For warning messages.
- `LogLevelError`: For error messages.
- `LogLevelSuccess`: For success messages.

## Using the Logger

Each `command.Command` instance has an embedded `log.DefaultLogger`. You can 
access the logger through the `Logger` field and use its methods to log messages:

- `Logger.Info(format string, a ...any)`
- `Logger.Warning(format string, a ...any)`
- `Logger.Error(format string, a ...any)`
- `Logger.Success(format string, a ...any)`

**Example:**

```go
func runExample(cmd *command.Command, rootFlags *command.RootFlags, args []string) error {
	cmd.Logger.Info("Starting example command...")
	// ...
	if someErrorOccurred {
		cmd.Logger.Error("An error has occurred: %v", someError)
		return someError
	}
	cmd.Logger.Success("Example command completed successfully.")
	return nil
}
```

## Log Message Formatting

Log messages are formatted with a timestamp (implicitly handled by 
`fmt.Printf`), a log level symbol, an optional component name (if the 
`DEBUG` environment variable is set to "1"), and the message itself.

## Log Level Colors

The library uses colors to distinguish between different log levels in the 
output:

- Info: Blue
- Warning: Yellow
- Error: Red
- Success: Green

## Aliases for Logging Functions

For convenience, the `command.Command` struct also includes alias functions 
that directly call the logger methods:

- `LogInfo(format string, a ...any)`
- `LogWarning(format string, a ...any)`
- `LogError(format string, a ...any)`
- `LogSuccess(format string, a ...any)`

You can use these directly within your command's `Run` function.

## Replacing the Default Logger

If you need more control over the logging process, you can replace the default 
logger with your own implementation. To do this, it's better to create a struct 
that has the same methods as the `log.DefaultLogger`: `Info`, `Warning`, 
`Error`, and `Success`.

Here's an example of how you might create a custom logger:

```go
package main

import (
	"fmt"

	"github.com/mirkobrombin/go-cli-builder/v1/command"
)

// CustomLogger is a custom logger implementation.
type CustomLogger struct {
	ComponentName string
}

// Info logs an informational message.
func (l *CustomLogger) Info(format string, a ...any) {
	fmt.Printf("[INFO - %s] %s\n", l.ComponentName, fmt.Sprintf(format, a...))
}

// Warning logs a warning message.
func (l *CustomLogger) Warning(format string, a ...any) {
	fmt.Printf("[WARNING - %s] %s\n", l.ComponentName, fmt.Sprintf(format, a...))
}

// Error logs an error message.
func (l *CustomLogger) Error(format string, a ...any) {
	fmt.Printf("[ERROR - %s] %s\n", l.ComponentName, fmt.Sprintf(format, a...))
}

// Success logs a success message.
func (l *CustomLogger) Success(format string, a ...any) {
	fmt.Printf("[SUCCESS - %s] %s\n", l.ComponentName, fmt.Sprintf(format, a...))
}

func main() {
	// Assuming you have a root command instance
	// rootCmd := root.NewRootCommand(...)

	// Create a custom logger
	customLogger := &CustomLogger{ComponentName: "MyApp"}

	// You can set the logger for the root command
	// rootCmd.Command.Logger = customLogger

	// Or for a specific subcommand
	exampleCmd := &command.Command{
		Name:        "example",
		Usage:       "example",
		Description: "An example command",
		Run: func(cmd *command.Command, rootFlags *command.RootFlags, argsstring) error {
			cmd.Logger.Info("This is an info message from the custom logger")
			return nil
		},
	}
	exampleCmd.SetupLogger("example") // This will set the default logger first
	exampleCmd.Logger = customLogger   // Override with the custom logger

	// rootCmd.AddCommand(exampleCmd)
	// ...
}
```

You can then assign an instance of your `CustomLogger` to the `Logger` field of 
either the `RootCommand` or individual `Command` instances. Remember to call 
`SetupLogger` first if you want the component name to be set initially.