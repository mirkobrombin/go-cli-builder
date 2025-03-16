package root

import (
	"fmt"

	"github.com/mirkobrombin/go-cli-builder/v1/command"
)

// addVersionCommand adds the version command to the root command.
func (rc *RootCommand) addVersionCommand() {
	versionCmd := &command.Command{
		Name:        "version",
		Usage:       "version",
		Description: "Print the application version",
		Run: func(cmd *command.Command, rootFlags *command.RootFlags, args []string) error {
			fmt.Println(rc.Version)
			return nil
		},
	}
	versionCmd.SetupLogger("version")
	rc.AddCommand(versionCmd)
}
