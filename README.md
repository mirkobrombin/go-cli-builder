# Go CLI Builder

A lightweight and flexible library for building command-line interfaces (CLIs) 
in Go. This library provides a simple and intuitive way to define commands, 
flags (including short names), aliases and more.

## Features

- **Simple Command Definition:** Easily define commands and subcommands with names, usage descriptions, and detailed descriptions.
- **Flag Handling:** Supports string, integer, and boolean flags with default values and the ability to allow arguments.
- **Short Flag Names:** Implements short names for flags (e.g., `-n` for `--name`).
- **Root Flags:** Allows defining flags at the root level that are accessible to all subcommands.
- **Environment Variable Integration:** Provides convenient functions to retrieve values from environment variables with default fallbacks.
- **Built-in Help Generation:** Automatically generates help messages for the root command and subcommands.
- **Customizable Logging:** Includes a basic logging system with different levels (info, warning, error, success).
- **Command Aliases:** Supports creating aliases for commands and subcommands.
- **Shell Completion:** Generates shell completion scripts for Bash, Zsh, and Fish.

## Getting Started

### Installation

```bash
go get github.com/mirkobrombin/go-cli-builder
```

### Basic Usage

```go
package main

import (
	"fmt"
	"os"

	"github.com/mirkobrombin/go-cli-builder/v1/root"
	"my-cli/commands"
)

func main() {
	rootCmd := root.NewRootCommand("mycli", "mycli [command]", "A simple CLI example", "1.0.0")

	rootCmd.AddBoolFlag("verbose", "v", "Enable verbose output", false, false)

	rootCmd.AddCommand(commands.NewAddCommand())
	rootCmd.AddCommand(commands.NewRemoveCommand())
	rootCmd.AddCommand(commands.NewListCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

Note, the `verbose` flag was registered at the root level, so it can be accessed
by all subcommands via the `rootFlags` parameter in the `Run` function, e.g.:

```go
func runSubcommand(cmd *command.Command, rootFlags *command.RootFlags, args []string) error {
    if rootFlags.GetBool("verbose") {
        cmd.Logger.Info("Verbose output enabled")
    }
    // ...
    return nil
}
```

For more detailed information, please refer to the documentation files in the 
[docs/](docs/) directory.

## Documentation

- [Command Management](docs/command.md)
- [Flag Handling](docs/flags.md)
- [Environment Variables](docs/environment_variables.md)
- [Logging](docs/logging.md)
- [Root Command](docs/root_command.md)
- [Aliases](docs/aliases.md)
- [Shell Completion](docs/completion.md)

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file 
for details.
