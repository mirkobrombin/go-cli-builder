package cli

import (
	"fmt"
	"io"
	"reflect"
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

`, fishQuote(name))
	if err != nil {
		return err
	}

	function := "__fish_" + name + "_using_command"
	if _, err := fmt.Fprintf(w, "function %s\n    set -l expected $argv[1]\n    set -l tokens (commandline -opc)\n    set -e tokens[1]\n    set -l path /\n    while true\n        switch \"$path\"\n", fishQuote(function)); err != nil {
		return err
	}
	if err := writeFishCommandPaths(w, a.RootNode, "/"); err != nil {
		return err
	}
	if _, err := io.WriteString(w, "        end\n    end\nend\n\n"); err != nil {
		return err
	}
	return writeFishCompletions(w, a.RootNode, name, function, "/")
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

// fishQuote keeps descriptions and shell arguments literal, including apostrophes.
func fishQuote(s string) string {
	return "'" + strings.NewReplacer(`\`, `\\`, "'", `\'`).Replace(s) + "'"
}

func fishFlagNames(node *parser.CommandNode) []string {
	names := make([]string, 0, len(node.Flags))
	for name := range node.Flags {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// writeFishCommandPaths parses each level separately so option values cannot
// be mistaken for subcommands.
func writeFishCommandPaths(w io.Writer, node *parser.CommandNode, path string) error {
	specs := []string{}
	for _, name := range fishFlagNames(node) {
		meta := node.Flags[name]
		spec := name
		if meta.Short != "" {
			spec = meta.Short + "/" + spec
		}
		if meta.Field.IsValid() && meta.Field.Kind() != reflect.Bool {
			spec += "="
		}
		specs = append(specs, fishQuote(spec))
	}
	if _, err := fmt.Fprintf(w, "            case %s\n                argparse -s %s -- $tokens 2>/dev/null; or return 1\n                set tokens $argv\n", fishQuote(path), strings.Join(specs, " ")); err != nil {
		return err
	}
	if len(node.Children) == 0 {
		_, err := io.WriteString(w, "                test \"$path\" = \"$expected\"\n                return\n")
		return err
	}
	if _, err := io.WriteString(w, "                if not set -q tokens[1]\n                    test \"$path\" = \"$expected\"\n                    return\n                end\n                switch $tokens[1]\n"); err != nil {
		return err
	}
	children := strings.Fields(listCommands(node, ""))
	for _, name := range children {
		child := node.Children[name]
		names := []string{fishQuote(name)}
		for _, alias := range child.Aliases {
			names = append(names, fishQuote(alias))
		}
		if _, err := fmt.Fprintf(w, "                    case %s\n                        set path %s\n", strings.Join(names, " "), fishQuote(strings.TrimSuffix(path, "/")+"/"+name)); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(w, "                    case '*'\n                        return 1\n                end\n                set -e tokens[1]\n"); err != nil {
		return err
	}
	for _, name := range children {
		if err := writeFishCommandPaths(w, node.Children[name], strings.TrimSuffix(path, "/")+"/"+name); err != nil {
			return err
		}
	}
	return nil
}

// writeFishCompletions recursively writes fish completion entries.
func writeFishCompletions(w io.Writer, node *parser.CommandNode, name, function, path string) error {
	condition := fishQuote(function) + " " + fishQuote(path)
	for _, flagName := range fishFlagNames(node) {
		meta := node.Flags[flagName]
		desc := meta.Description
		if desc == "" {
			desc = flagName
		}
		short := ""
		if meta.Short != "" {
			short = "-s " + fishQuote(meta.Short)
		}
		_, err := fmt.Fprintf(w, "complete -c %s -n %s -l %s %s -d %s\n",
			fishQuote(name), fishQuote(condition), fishQuote(flagName), short, fishQuote(desc))
		if err != nil {
			return err
		}
	}
	for _, childName := range strings.Fields(listCommands(node, "")) {
		child := node.Children[childName]
		for _, command := range append([]string{childName}, child.Aliases...) {
			if _, err := fmt.Fprintf(w, "complete -c %s -n %s -a %s -d %s\n",
				fishQuote(name), fishQuote(condition), fishQuote(command), fishQuote(child.Description)); err != nil {
				return err
			}
		}
		if err := writeFishCompletions(w, child, name, function, strings.TrimSuffix(path, "/")+"/"+childName); err != nil {
			return err
		}
	}
	return nil
}
