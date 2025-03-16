package root

import (
	"github.com/mirkobrombin/go-cli-builder/v1/command"
)

// RootCommand represents the root command. Can be assigned while creating
// a new root command with NewRootCommand.
type RootCommand struct {
	command.Command

	// Version is the version of the application.
	Version string

	// Commands maps command names to commands.
	Commands map[string]*command.Command
}
