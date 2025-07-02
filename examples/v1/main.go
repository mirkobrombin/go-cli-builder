package main

import (
	"fmt"
	"os"

	"github.com/mirkobrombin/go-cli-builder/examples/v1/commands"
	"github.com/mirkobrombin/go-cli-builder/v1/root"
)

// main is the entry point of the CLI application.
func main() {
	// Create a new root command for the CLI.
	rootCmd := root.NewRootCommand("mycli", "mycli [command]", "A simple CLI example", "1.0.0")

	// Add flags to the root command.
	rootCmd.AddBoolFlag("verbose", "v", "Enable verbose output", false, false, true)

	// Add subcommands to the root command.
	rootCmd.AddCommand(commands.NewAddCommand())
	rootCmd.AddCommand(commands.NewRemoveCommand())
	rootCmd.AddCommand(commands.NewListCommand())

	// Execute the root command and handle any errors.
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
