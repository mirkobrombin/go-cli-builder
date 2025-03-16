package commands

import (
	"fmt"
	"strings"

	"github.com/mirkobrombin/go-cli-builder/examples/v1/core"
	"github.com/mirkobrombin/go-cli-builder/v1/command"
)

// NewListCommand creates a new 'list' command.
//
// Returns:
//   - A pointer to a new command.Command representing the 'list' command.
func NewListCommand() *command.Command {
	cmd := &command.Command{
		Name:        "list",
		Usage:       "list",
		Description: "List all items",
		Run:         runList,
	}
	return cmd
}

// runList executes the 'list' command logic.
// It loads data from the JSON file and prints all items.
//
// Parameters:
//   - cmd: A pointer to the command.Command that is being executed.
//   - args: A slice of strings representing the command-line arguments.
//
// Returns:
//   - An error if the command execution fails, or nil if successful.
func runList(cmd *command.Command, rootFlags *command.RootFlags, args []string) error {
	if rootFlags.GetBool("verbose") {
		cmd.Logger.Info("You see this because the 'verbose' flag is set in the root command")
	}

	data, err := core.LoadData()
	if err != nil {
		if rootFlags.GetBool("verbose") {
			cmd.Logger.Error("Error loading data")
		}
		return err
	}

	if len(data.Items) == 0 {
		cmd.Logger.Info("No items found")
		return nil
	}

	var sb strings.Builder
	sb.WriteString("Items:\n")
	for _, item := range data.Items {
		sb.WriteString(fmt.Sprintf("· %s\n", item))
	}
	cmd.Logger.Success(sb.String())

	return nil
}
