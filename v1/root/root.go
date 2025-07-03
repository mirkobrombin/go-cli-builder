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
	// Required flags check
	setFlags := make(map[string]bool)
	var requiredFlags []string

	// Parse root flags
	if len(os.Args) > 1 {
		if err := rc.Flags.Parse(os.Args[1:]); err != nil {
			return err
		}

		// Check for required flags on the root command
		rc.Flags.Visit(func(f *flag.Flag) {
			setFlags[f.Name] = true
		})

		for _, requiredFlag := range rc.RequiredFlags {
			if _, isSet := setFlags[requiredFlag]; !isSet {
				requiredFlags = append(requiredFlags, "'-"+requiredFlag+"'")
			}
		}
	}

	args := rc.Flags.Args()
	if len(args) == 0 {
		rc.PrintHelp()
		return nil
	}

	var cmd *command.Command
	currentLevelCmds := rc.SubCommands
	var commandsTraversed int
	// var hasSubCommands bool

	// Find the command to execute by traversing the command tree
	for i, arg := range args {
		var foundCmd *command.Command
		for _, c := range currentLevelCmds {
			if c.Name == arg {
				foundCmd = c
				break
			}
		}

		if foundCmd != nil {
			cmd = foundCmd
			// hasSubCommands = len(cmd.SubCommands) > 0
			currentLevelCmds = cmd.SubCommands
			commandsTraversed = i + 1
		} else {
			break
		}
	}

	if cmd == nil {
		rc.PrintHelp()
		return nil
	}

	cmdArgs := args[commandsTraversed:]

	// If the command has subcommands and no arguments are provided, show help
	if len(cmd.SubCommands) > 0 && len(cmdArgs) == 0 {
		cmd.PrintCommandHelp()
		return nil
	}

	// Handle help flag for the found subcommand
	for _, arg := range cmdArgs {
		if arg == "-h" || arg == "--help" {
			cmd.PrintCommandHelp()
			return nil
		}
	}

	var remainingArgs []string

	// Parse flags for the subcommand
	if cmd.Flags != nil {
		expandedArgs := make([]string, 0, len(cmdArgs))
		for _, arg := range cmdArgs {
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

		// Check for required flags on the subcommand
		cmd.Flags.Visit(func(f *flag.Flag) {
			setFlags[f.Name] = true
		})

		for _, requiredFlag := range cmd.RequiredFlags {
			if _, isSet := setFlags[requiredFlag]; !isSet {
				requiredFlags = append(requiredFlags, "'-"+requiredFlag+"'")
			}
		}

		remainingArgs = cmd.Flags.Args()
	} else {
		remainingArgs = cmdArgs
	}

	// Check for any missing required flags (both root and subcommand)
	// add a subcommand to a subcommand
	if len(requiredFlags) > 0 {
		cmd.Logger.Error("Missing required flags: %s", strings.Join(requiredFlags, ", "))
		cmd.PrintCommandHelp()
		return nil
	}

	if cmd.Run == nil {
		err := fmt.Errorf("no main `Run` function defined for command, can't call any hooks (BeforeRun, Run, AfterRun) '%s'", cmd.Name)
		cmd.Logger.Error("%v", err)
		return err
	}

	parsedRootFlags := &command.RootFlags{FlagSet: rc.Flags}

	if cmd.BeforeRun != nil {
		if err := cmd.BeforeRun(cmd, parsedRootFlags, remainingArgs); err != nil {
			cmd.Logger.Error("Error running before main command: %v", err)
			return err
		}
	}

	if err := cmd.Run(cmd, parsedRootFlags, remainingArgs); err != nil {
		cmd.Logger.Error("Error running main command: %v", err)
		return err
	}

	if cmd.AfterRun != nil {
		if err := cmd.AfterRun(cmd, parsedRootFlags, remainingArgs); err != nil {
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

// AddAlias adds an alias for a command or subcommand chain of any depth.
//
// Parameters:
//   - alias: The name of the alias.
//   - targets: The target command and optional subcommands in sequence (e.g., "cmd", "subcmd", "subsubcmd").
func (rc *RootCommand) AddAlias(alias string, targets ...string) {
	if len(targets) == 0 {
		fmt.Fprintf(os.Stderr, "No target command specified for alias '%s'\n", alias)
		return
	}

	// Check if the first command exists
	if _, ok := rc.Commands[targets[0]]; !ok {
		fmt.Fprintf(os.Stderr, "Target command '%s' not found for alias '%s'\n", targets[0], alias)
		return
	}

	// Build description based on targets depth
	var description string
	if len(targets) == 1 {
		description = fmt.Sprintf("Alias for '%s'", targets[0])
	} else {
		description = fmt.Sprintf("Alias for '%s'", strings.Join(targets, " "))
	}

	// Create alias command
	aliasCmd := &command.Command{
		Name:        alias,
		Usage:       alias,
		Description: description,
		Run: func(cmd *command.Command, rootFlags *command.RootFlags, args []string) error {
			newArgs := append([]string{}, targets...)
			newArgs = append(newArgs, args...)
			return utils.Reexec(newArgs)
		},
	}

	aliasCmd.SetupLogger(alias)
	rc.AddCommand(aliasCmd)
}

// AddAliasSubCommand adds an alias for a subcommand of a command.
// Deprecated: Use AddAlias instead. This function exists for backward compatibility.
//
// Parameters:
//   - alias: The name of the alias.
//   - targetCmd: The name of the target command.
//   - targetSubCmd: The name of the target subcommand.
//   - targets: Optional additional targets for deeper nesting.
func (rc *RootCommand) AddAliasSubCommand(alias, targetCmd, targetSubCmd string, targets ...string) {
	allTargets := append([]string{targetCmd, targetSubCmd}, targets...)
	rc.AddAlias(alias, allTargets...)
}

// AddAliasCommand adds an alias for a command.
// Deprecated: Use AddAlias instead. This function exists for backward compatibility.
//
// Parameters:
//   - alias: The name of the alias.
//   - targetCmd: The name of the target command.
func (rc *RootCommand) AddAliasCommand(alias, targetCmd string) {
	rc.AddAlias(alias, targetCmd)
}
