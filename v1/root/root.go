package root

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mirkobrombin/go-cli-builder/v1/command"
	"github.com/mirkobrombin/go-cli-builder/v1/log"
	"github.com/mirkobrombin/go-cli-builder/v1/utils"
)

// NewRootCommand creates a new root command.
//
// Parameters:
//   - name: The name of the root command.
//   - usage: The usage of the root command.
//   - description: The description of the root command.
//   - version: The version of the application.
//
// Returns:
//   - A pointer to a new RootCommand.
func NewRootCommand(name, usage, description string, version string) *RootCommand {
	rc := &RootCommand{
		Version: version,
		Command: command.Command{
			Name:        name,
			Usage:       usage,
			Description: description,
			Flags:       flag.NewFlagSet(name, flag.ExitOnError),
			SubCommands: []*command.Command{},
			ArgFlags:    make(map[string]bool),
			Logger:      &log.DefaultLogger{ComponentName: name},
		},
		Commands: make(map[string]*command.Command),
	}
	// Adds the built-in commands (version and completion)
	rc.addBuiltInCommands()
	return rc
}

// AddCommand adds a subcommand.
//
// Parameters:
//   - cmd: The subcommand to add.
func (rc *RootCommand) AddCommand(cmd *command.Command) {
	cmd.Logger = &log.DefaultLogger{ComponentName: cmd.Name}
	rc.SubCommands = append(rc.SubCommands, cmd)
	rc.Commands[cmd.Name] = cmd
}

// Execute runs the main command.
//
// Returns:
//   - An error, if an error occurs during execution.
func (rc *RootCommand) Execute() error {
	// Parse root flags
	if len(os.Args) > 1 {
		if err := rc.Flags.Parse(os.Args[1:]); err != nil {
			return err
		}
	}

	args := rc.Flags.Args()
	if len(args) < 1 {
		rc.PrintHelp()
		return nil
	}

	cmdName := args[0]
	cmd, ok := rc.Commands[cmdName]
	if !ok {
		rc.PrintHelp()
		return nil
	}

	// Handle subcommand aliases
	parsedRootFlags := &command.RootFlags{FlagSet: rc.Flags}
	if len(args) >= 2 {
		subCmdName := args[1]
		if aliasCmd, ok := rc.Commands[cmdName]; ok {
			if aliasCmd.Description == fmt.Sprintf("Alias for '%s %s <branch>'", cmdName, subCmdName) {
				if subCmd, ok := rc.Commands[subCmdName]; ok {
					return subCmd.Run(subCmd, parsedRootFlags, os.Args[3:])
				}
				return fmt.Errorf("subcommand '%s' not found", subCmdName)
			}
		}
	}

	// Handle help flag, we also avoid it to be parsed by the subcommand
	// as a positional argument
	for _, arg := range args[1:] {
		if arg == "-h" || arg == "--help" {
			cmd.PrintCommandHelp()
			return nil
		}
	}

	var remainingArgs []string

	// Parse flags if any
	if cmd.Flags != nil {
		expandedArgs := make([]string, 0, len(args)-1)
		for _, arg := range args[1:] {
			if strings.HasPrefix(arg, "-") && !strings.HasPrefix(arg, "--") && len(arg) == 2 {
				shortName := strings.TrimPrefix(arg, "-")
				if longName, ok := cmd.ShortFlagMap[shortName]; ok {
					expandedArgs = append(expandedArgs, fmt.Sprintf("--%s", longName))
				} else {
					expandedArgs = append(expandedArgs, arg)
				}
			} else {
				expandedArgs = append(expandedArgs, arg)
			}
		}

		if err := cmd.Flags.Parse(expandedArgs); err != nil {
			return err
		}

		// Handle positional arguments
		remainingArgs = cmd.Flags.Args()
		argIndex := 0

		// Handle flags with arguments
		for _, arg := range remainingArgs {
			if strings.HasPrefix(arg, "-") {
				flagName := strings.TrimPrefix(arg, "-")
				if strings.HasPrefix(flagName, "-") { // Long flag
					flagName = strings.TrimPrefix(flagName, "-")
				} else if longName, ok := cmd.ShortFlagMap[flagName]; ok { // Short flag
					flagName = longName
				}
				if allowArg, ok := cmd.ArgFlags[flagName]; ok && allowArg {
					if argIndex+1 < len(remainingArgs) {
						cmd.Flags.Set(flagName, remainingArgs[argIndex+1])
						argIndex += 2
					}
				}
			} else {
				// Handle flags without arguments
				for flagName, allowArg := range cmd.ArgFlags {
					if allowArg && argIndex < len(remainingArgs) {
						cmd.Flags.Set(flagName, remainingArgs[argIndex])
						argIndex++
					}
				}
			}
		}
	}

	var finalArgs []string

	if len(remainingArgs) > 0 {
		finalArgs = remainingArgs
	} else {
		finalArgs = args[1:]
	}

	if cmd.Run == nil {
		err := fmt.Errorf("no main `Run` function defined for command, can't call any hooks (BeforeRun, Run, AfterRun) '%s'", cmd.Name)
		cmd.Logger.Error("%v", err)
		return err
	}

	if cmd.BeforeRun != nil {
		if err := cmd.BeforeRun(cmd, parsedRootFlags, finalArgs); err != nil {
			cmd.Logger.Error("Error running before main command: %v", err)
			return err
		}
	}

	if err := cmd.Run(cmd, parsedRootFlags, finalArgs); err != nil {
		cmd.Logger.Error("Error running main command: %v", err)
		return err
	}

	if cmd.AfterRun != nil {
		if err := cmd.AfterRun(cmd, parsedRootFlags, finalArgs); err != nil {
			cmd.Logger.Error("Error running after main command: %v", err)
			return err
		}
	}

	return nil
}

// PrintHelp prints the help message for the root command.
func (rc *RootCommand) PrintHelp() {
	fmt.Printf("%s - %s\n\n", rc.Name, rc.Description)
	if rc.Version != "" {
		fmt.Printf("Version: %s\n\n", rc.Version)
	}
	fmt.Printf("Usage: %s\n\n", rc.Usage)

	fmt.Println("Commands:")
	maxNameLength := 0
	for _, cmd := range rc.SubCommands {
		if len(cmd.Name) > maxNameLength {
			maxNameLength = len(cmd.Name)
		}
	}
	for _, cmd := range rc.SubCommands {
		fmt.Printf("  %-*s  %s\n", maxNameLength, cmd.Name, cmd.Description)
	}
	if rc.Flags != nil {
		fmt.Println("\nFlags:")
		rc.Flags.VisitAll(func(f *flag.Flag) {
			fmt.Printf("  -%s: %s\n", f.Name, f.Usage)
		})
	}
}

// AddAliasSubCommand adds an alias for a subcommand of a command.
//
// Parameters:
//   - alias: The name of the alias.
//   - targetCmd: The name of the target command.
//   - targetSubCmd: The name of the target subcommand.
func (rc *RootCommand) AddAliasSubCommand(alias, targetCmd, targetSubCmd string) {
	if _, ok := rc.Commands[targetCmd]; ok {
		aliasCmd := &command.Command{
			Name:        alias,
			Usage:       fmt.Sprintf("%s <branch>", alias),
			Description: fmt.Sprintf("Alias for '%s %s <branch>'", targetCmd, targetSubCmd),
			Run: func(cmd *command.Command, rootFlags *command.RootFlags, args []string) error {
				newArgs := append([]string{targetCmd, targetSubCmd}, args...)
				return utils.Reexec(newArgs)
			},
		}
		aliasCmd.SetupLogger(alias)
		rc.AddCommand(aliasCmd)
	} else {
		fmt.Fprintf(os.Stderr, "Target command '%s' not found for alias '%s'\n", targetCmd, alias)
	}
}

// AddAliasCommand adds an alias for a command.
//
// Parameters:
//   - alias: The name of the alias.
//   - targetCmd: The name of the target command.
func (rc *RootCommand) AddAliasCommand(alias, targetCmd string) {
	if _, ok := rc.Commands[targetCmd]; ok {
		aliasCmd := &command.Command{
			Name:        alias,
			Usage:       alias,
			Description: fmt.Sprintf("Alias for '%s'", targetCmd),
			Run: func(cmd *command.Command, rootFlags *command.RootFlags, args []string) error {
				newArgs := append([]string{targetCmd}, args...)
				return utils.Reexec(newArgs)
			},
		}
		aliasCmd.SetupLogger(alias)
		rc.AddCommand(aliasCmd)
	} else {
		fmt.Fprintf(os.Stderr, "Target command '%s' not found for alias '%s'\n", targetCmd, alias)
	}
}
