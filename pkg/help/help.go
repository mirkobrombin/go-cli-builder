package help

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mirkobrombin/go-cli-builder/v2/pkg/parser"
)

// Translator is a function that translates a key.
type Translator func(string) string

// GenerateHelp generates a formatted help string for a command node.
//
// Example:
//
//	helpText := help.GenerateHelp(rootNode, nil)
//	fmt.Println(helpText)
func GenerateHelp(node *parser.CommandNode, tr Translator) string {
	var sb strings.Builder

	t := func(s string) string {
		if tr != nil && strings.HasPrefix(s, "pr:") {
			return tr(strings.TrimPrefix(s, "pr:"))
		}
		return s
	}

	fmt.Fprintf(&sb, "Usage: %s [flags]", node.Name)
	if len(node.Children) > 0 {
		fmt.Fprintf(&sb, " [command]")
	}
	if len(node.Args) > 0 {
		for _, arg := range node.Args {
			if arg.IsGreedy {
				fmt.Fprintf(&sb, " [%s...]", t(arg.Description))
			} else {
				fmt.Fprintf(&sb, " [%s]", t(arg.Description))
			}
		}
	}
	fmt.Fprint(&sb, "\n\n")

	if node.Description != "" {
		fmt.Fprintf(&sb, "%s\n\n", t(node.Description))
	}

	if len(node.Children) > 0 {
		fmt.Fprint(&sb, "Commands:\n")
		cmdNames := make([]string, 0, len(node.Children))
		for name := range node.Children {
			if node.Children[name].Name == name {
				cmdNames = append(cmdNames, name)
			}
		}
		sort.Strings(cmdNames)

		for _, name := range cmdNames {
			child := node.Children[name]
			aliases := ""
			if len(child.Aliases) > 0 {
				aliases = fmt.Sprintf(" (aliases: %s)", strings.Join(child.Aliases, ", "))
			}
			fmt.Fprintf(&sb, "  %-15s %s%s\n", name, t(child.Description), aliases)
		}
		fmt.Fprint(&sb, "\n")
	}

	if len(node.Flags) > 0 {
		fmt.Fprint(&sb, "Flags:\n")

		flagNames := make([]string, 0, len(node.Flags))
		for name := range node.Flags {
			flagNames = append(flagNames, name)
		}
		sort.Strings(flagNames)

		for _, name := range flagNames {
			meta := node.Flags[name]
			short := ""
			if meta.Short != "" {
				short = fmt.Sprintf("-%s, ", meta.Short)
			}

			details := []string{}
			if meta.Env != "" {
				details = append(details, fmt.Sprintf("env: %s", meta.Env))
			}
			if meta.Default != "" {
				details = append(details, fmt.Sprintf("default: %s", meta.Default))
			}
			if meta.Required {
				details = append(details, "required")
			}

			detailStr := ""
			if len(details) > 0 {
				detailStr = fmt.Sprintf(" (%s)", strings.Join(details, ", "))
			}

			fmt.Fprintf(&sb, "  %s--%-12s %s%s\n", short, name, t(meta.Description), detailStr)
		}
	}

	return sb.String()
}
