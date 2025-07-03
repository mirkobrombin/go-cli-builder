package commands

import (
	"fmt"

	"github.com/mirkobrombin/go-cli-builder/examples/v1/core"
	"github.com/mirkobrombin/go-cli-builder/v1/command"
)

// NewAddCommand creates a new 'add' command.
//
// Returns:
//   - A pointer to a new command.Command representing the 'add' command.
func NewAddCommand() *command.Command {
	cmd := &command.Command{
		Name:        "add",
		Usage:       "add <item>",
		Description: "Add a new item to the list",
		BeforeRun:   runBeforeAdd,
		AfterRun:    runAfterAdd,
		Run:         runAdd,
	}
	cmd.AddFlag("name", "n", "Name of the item to add", "", true, true)
	return cmd
}

// runBeforeAdd executes the 'runBefore' command logic.
// It shows the present data before of the addition of the item

// Parameters:
//   - cmd: A pointer to the command.Command that is being executed.
//   - args: A slice of strings representing the command-line arguments.
//
// Returns:
//   - An error if the command execution fails, or nil if successful.
func runBeforeAdd(cmd *command.Command, rootFlags *command.RootFlags, args []string) error {
	data, err := core.LoadData()
	if err != nil {
		if rootFlags.GetBool("verbose") {
			cmd.Logger.Error("Error loading data")
		}
		return err
	}
	if len(data.Items) >= 3 {
		cmd.Logger.Info("Last 3 items before addition: %s", data.Items[len(data.Items)-3:])
	}
	return nil
}

// runAfterAdd executes the 'runAfter' command logic.
// It shows the present data after of the addition of the item

// Parameters:
//   - cmd: A pointer to the command.Command that is being executed.
//   - args: A slice of strings representing the command-line arguments.
//
// Returns:
//   - An error if the command execution fails, or nil if successful.
func runAfterAdd(cmd *command.Command, rootFlags *command.RootFlags, args []string) error {
	data, err := core.LoadData()
	if err != nil {
		if rootFlags.GetBool("verbose") {
			cmd.Logger.Error("Error loading data")
		}
		return err
	}
	if len(data.Items) >= 3 {
		cmd.Logger.Info("Last 3 items after addition: %s", data.Items[len(data.Items)-3:])
	}
	return nil
}

// runAdd executes the 'add' command logic.
// It loads data from the JSON file, appends the new item, and saves the updated data.
//
// Parameters:
//   - cmd: A pointer to the command.Command that is being executed.
//   - args: A slice of strings representing the command-line arguments.
//
// Returns:
//   - An error if the command execution fails, or nil if successful.
func runAdd(cmd *command.Command, rootFlags *command.RootFlags, args []string) error {
	if rootFlags.GetBool("verbose") {
		cmd.Logger.Info("You see this because the 'verbose' flag is set in the root command")
	}

	data, err := core.LoadData()
	if err != nil {
		cmd.Logger.Error("Error loading data")
		return err
	}

	name := cmd.GetFlagString("name")

	if name == "" {
		cmd.Logger.Error("Missing item name")
		return fmt.Errorf("missing item name")
	}

	data.Items = append(data.Items, name)

	if err := core.SaveData(data); err != nil {
		cmd.Logger.Error("Error adding item: %s", name)
		return err
	}

	cmd.Logger.Success(fmt.Sprintf("Added item: %s", name))
	return nil
}
