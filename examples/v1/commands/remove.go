package commands

import (
	"fmt"

	"github.com/mirkobrombin/go-cli-builder/examples/v1/core"
	"github.com/mirkobrombin/go-cli-builder/v1/command"
)

// NewRemoveCommand creates a new 'remove' command.
//
// Returns:
//   - A pointer to a new command.Command representing the 'remove' command.
func NewRemoveCommand() *command.Command {
	cmd := &command.Command{
		Name:        "remove",
		Usage:       "remove <item>",
		Description: "Remove an item from the list",
		Run:         runRemove,
	}
	cmd.AddFlag("name", "n", "Name of the item to add", "", true)
	return cmd
}

// runRemove executes the 'remove' command logic.
// It loads data from the JSON file, removes the specified item, and saves the updated data.
//
// Parameters:
//   - cmd: A pointer to the command.Command that is being executed.
//   - args: A slice of strings representing the command-line arguments.
//
// Returns:
//   - An error if the command execution fails, or nil if successful.
func runRemove(cmd *command.Command, rootFlags *command.RootFlags, args []string) error {
	if rootFlags.GetBool("verbose") {
		cmd.Logger.Info("You see this because the 'verbose' flag is set in the root command")
	}

	name := cmd.GetFlagString("name")
	if name == "" {
		cmd.Logger.Error("Missing item name")
		return fmt.Errorf("missing item name")
	}

	data, err := core.LoadData()
	if err != nil {
		cmd.Logger.Error("Error loading data")
		return err
	}

	var updatedItems []string
	for _, existingItem := range data.Items {
		if existingItem != name {
			updatedItems = append(updatedItems, existingItem)
		}
	}
	data.Items = updatedItems

	if err := core.SaveData(data); err != nil {
		cmd.Logger.Error("Error removing item: %s", name)
		return err
	}

	cmd.Logger.Success(fmt.Sprintf("Removed item: %s", name))
	return nil
}
