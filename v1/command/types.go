package command

import (
	"flag"

	"github.com/mirkobrombin/go-cli-builder/v1/log"
)

type RunFunc func(cmd *Command, rootFlags *RootFlags, args []string) error

// Command represents a CLI command.
type Command struct {
	// Name is the name of the command.
	Name string

	// Usage is the short usage description of the command.
	Usage string

	// Description is the detailed description of the command.
	Description string

	// Flags are the command-line flags for the command.
	Flags *flag.FlagSet

	// Run is the function to execute when the command is invoked.
	Run RunFunc

	// SubCommands are the subcommands of the command.
	SubCommands []*Command

	// ArgFlags maps flag names to a boolean indicating if the flag allows an argument.
	ArgFlags map[string]bool

	// ShortFlagMap maps short flag names to their corresponding long flag names.
	ShortFlagMap map[string]string

	// Logger is an instance of the active logger.
	Logger *log.DefaultLogger

	// LogInfo is the function to log informational messages.
	LogInfo func(format string, a ...any)

	// LogWarning is the function to log warning messages.
	LogWarning func(format string, a ...any)

	// LogError is the function to log error messages.
	LogError func(format string, a ...any)

	// LogSuccess is the function to log success messages.
	LogSuccess func(format string, a ...any)
}

// RootFlags represents the parsed flags of the root command.
type RootFlags struct {
	*flag.FlagSet
}
