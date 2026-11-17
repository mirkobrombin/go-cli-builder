package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/mirkobrombin/go-cli-builder/v3/pkg/parser"
)

// GenBashCompletion writes a bash completion script for the app.
func (a *App) GenBashCompletion(w io.Writer) error {
	name := a.RootNode.Name
	if name == "" {
		name = "app"
	}

	_, err := fmt.Fprintf(w, `_%s_completions()
{
    local cur prev words cword
    _init_completion || return

    COMPREPLY=($(compgen -W "%s" -- "$cur"))
}

complete -F _%s_completions %s
`, name, listCommands(a.RootNode, ""), name, name)
	return err
}

// GenZshCompletion writes a zsh completion script for the app.
func (a *App) GenZshCompletion(w io.Writer) error {
	name := a.RootNode.Name
	if name == "" {
		name = "app"
	}

	cmds := listCommands(a.RootNode, "")
	_, err := fmt.Fprintf(w, `#compdef %s

_%s() {
    local -a subcommands
    subcommands=(%s)
    _describe 'command' subcommands
}

_%s
`, name, name, cmds, name)
	return err
}

// GenFishCompletion writes a fish completion script for the app.
func (a *App) GenFishCompletion(w io.Writer) error {
	name := a.RootNode.Name
	if name == "" {
		name = "app"
	}

	_, err := fmt.Fprintf(w, `complete -c %s -f

`, name)
	if err != nil {
		return err
	}

	return writeFishCompletions(w, a.RootNode, name, []string{})
}

// listCommands recursively lists commands space-separated for bash completion.
func listCommands(node *parser.CommandNode, prefix string) string {
	names := make([]string, 0, len(node.Children))
	for name, child := range node.Children {
		if child.Name == name {
			fullName := prefix + name
			names = append(names, fullName)
		}
	}
	sort.Strings(names)
	return strings.Join(names, " ")
}

// writeFishCompletions recursively writes fish completion entries.
func writeFishCompletions(w io.Writer, node *parser.CommandNode, name string, parentNames []string) error {
	for flagName, meta := range node.Flags {
		desc := meta.Description
		if desc == "" {
			desc = flagName
		}
		short := ""
		if meta.Short != "" {
			short = fmt.Sprintf("-s %s", meta.Short)
		}
		conditions := fishSubcommandConditions(parentNames, map[string]*parser.CommandNode{})
		_, err := fmt.Fprintf(w, "complete -c %s %s -l %s %s -d '%s'\n",
			name, strings.Join(conditions, " "), flagName, short, escapeFishSingleQuoted(desc))
		if err != nil {
			return err
		}
	}

	for childName, child := range node.Children {
		if child.Name != childName {
			continue
		}

		newParentNames := append(parentNames, childName)

		conditions := []string{"-n '__fish_use_subcommand'"}
		if len(parentNames) > 0 {
			conditions = fishSubcommandConditions(parentNames, node.Children)
		}

		_, err := fmt.Fprintf(w, "complete -c %s %s -a %s -d '%s'\n",
			name, strings.Join(conditions, " "), childName, escapeFishSingleQuoted(child.Description))
		if err != nil {
			return err
		}

		if err := writeFishCompletions(w, child, name, newParentNames); err != nil {
			return err
		}
	}
	return nil
}

// escapeFishSingleQuoted makes s safe to place between single quotes in
// fish-shell
func escapeFishSingleQuoted(s string) string {
	var fishSingleQuoteEscaper = strings.NewReplacer(`\`, `\\`, `'`, `\'`)
	return fishSingleQuoteEscaper.Replace(s)
}

// fishSubcommandConditions builds fish's complete -n conditions for a
// subcommand given its parents and children
func fishSubcommandConditions(parents []string, children map[string]*parser.CommandNode) []string {
	conditions := []string{}
	for _, parent := range parents {
		conditions = append(conditions, fmt.Sprintf("-n '__fish_seen_subcommand_from %s'", parent))
	}

	childNames := []string{}
	for childName, child := range children {
		if childName != child.Name {
			continue
		}
		childNames = append(childNames, childName)
	}
	if len(childNames) > 0 {
		conditions = append(conditions, fmt.Sprintf("-n 'not __fish_seen_subcommand_from %s'",
			strings.Join(childNames, " ")))
	}
	return conditions
}
