# Command Aliases

The Go CLI Builder library allows you to define aliases for commands and 
subcommands. This can provide a more convenient or intuitive way for users to 
interact with your CLI.

## Adding Command Aliases

You can add aliases using the following methods of the `RootCommand`:

- `AddAliasCommand(alias, targetCmd string)`: Creates an alias for an entire command. When the `alias` is used, the `targetCmd` will be executed.
- `AddAliasSubCommand(alias, targetCmd, targetSubCmd string)`: Creates an alias for a specific subcommand of a command. When the `alias` followed by `<branch>` is used, the `targetCmd` and `targetSubCmd` will be executed with the provided branch.

**Example:**

```go
package main

import (
	"fmt"
	"os"

	"github.com/mirkobrombin/go-cli-builder/v1/root"
	"github.com/mirkobrombin/go-cli-builder/examples/v1/commands"
)

func main() {
	rootCmd := root.NewRootCommand("mycli", "mycli [command]", "A simple CLI example", "1.0.0")

	rootCmd.AddCommand(commands.NewAddCommand())
	rootCmd.AddCommand(commands.NewRemoveCommand())
	rootCmd.AddCommand(commands.NewListCommand())

	// Assuming you have a subcommand 'switch' under a command 'mode' and you
	// want to create an alias 'sw' for it.
	rootCmd.AddAliasSubCommand("sw", "mode", "switch")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

In this example, users can now use `mycli sw` instead of `mycli mode switch`.

**Note:** Aliases are implemented by re-executing the CLI with the target 
command and arguments.