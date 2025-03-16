package root

// addBuiltInCommands add all the built-in commands to the root command.
func (rc *RootCommand) addBuiltInCommands() {
	rc.addVersionCommand()
	rc.addCompletionCommand()
}

// getCommandNames returns a slice with the names of all commands.
func (rc *RootCommand) getCommandNames() []string {
	var names []string
	for _, cmd := range rc.SubCommands {
		names = append(names, cmd.Name)
	}
	return names
}
