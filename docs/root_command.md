# Root Command

The `RootCommand` is the entry point of your CLI application. It manages 
subcommands, root-level flags, and the overall execution flow.

## Creating a Root Command

You create a new `RootCommand` using the `root.NewRootCommand()` function:

```go
package main

import (
	"github.com/mirkobrombin/go-cli-builder/v1/root"
)

func main() {
	rootCmd := root.NewRootCommand("mycli", "mycli [command]", "A simple CLI example", "1.0.0")
	// ...
}
```

Parameters:

- `name`: The name of the root command (e.g., "mycli").
- `usage`: The usage description for the root command (e.g., "mycli [command]").
- `description`: A detailed description of the application.
- `version`: The version of the application.

## Adding Commands

You add subcommands to the `RootCommand` using the `AddCommand()` method:

```go
rootCmd.AddCommand(commands.NewAddCommand())
rootCmd.AddCommand(commands.NewRemoveCommand())
```

## Root Flags

You can add flags that are applicable to the entire application at the root 
level using the `AddFlag()`, `AddIntFlag()`, and `AddBoolFlag()` methods of 
the `RootCommand`. These flags are parsed before any subcommand is executed, 
and their values can be accessed in the subcommand's `Run` function via the 
`rootFlags` parameter.

## Executing the Root Command

The `Execute()` method of the `RootCommand` starts the CLI application. It 
parses the command-line arguments, finds the appropriate command to execute, 
and calls its `Run` function.

```go
if err := rootCmd.Execute(); err != nil {
	// Handle errors
}
```

## Printing Help

The `PrintHelp()` method displays the help message for the root command, 
including its usage, description, available commands, and root flags. This is 
automatically called if no command is provided or if the user requests help 
for the root command.

## Adding Built-in Commands

The `NewRootCommand()` function automatically adds two built-in commands:

- `version`: Prints the application version.
- `completion`: Generates shell completion scripts.