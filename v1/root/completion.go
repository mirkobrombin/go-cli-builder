package root

import (
	"fmt"
	"strings"

	"github.com/mirkobrombin/go-cli-builder/v1/command"
)

// addCompletionCommand adds the 'completion' command to the root command.
func (rc *RootCommand) addCompletionCommand() {
	completionCmd := &command.Command{
		Name:        "completion",
		Usage:       "completion [bash|zsh|fish]",
		Description: "Generates shell completion scripts",
		Run:         rc.runCompletion,
	}
	completionCmd.SetupLogger("completion")
	rc.AddCommand(completionCmd)
}

// runCompletion generates the completion script for the specified shell.
func (rc *RootCommand) runCompletion(cmd *command.Command, rootFlags *command.RootFlags, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("shell type required (bash, zsh, fish)")
	}

	shell := args[0]
	switch shell {
	case "bash":
		return rc.genBashCompletion()
	case "zsh":
		return rc.genZshCompletion()
	case "fish":
		return rc.genFishCompletion()
	default:
		return fmt.Errorf("unsupported shell type: %s", shell)
	}
}

// genBashCompletion generates the bash completion script.
func (rc *RootCommand) genBashCompletion() error {
	fmt.Printf(`
_%s_completion()
{
	local cur prev opts
	COMPREPLY=()
	cur="${COMP_WORDS[COMP_CWORD]}"
	prev="${COMP_WORDS[COMP_PREV_CWORD]}"
	opts="%s"

	if [[ "${cur}" == "-" ]]; then
			COMPREPLY=( $(compgen -W "${opts}" -- "${cur}") )
			return 0
	fi

	local completions=()
	for opt in ${opts}; do
			if [[ "${opt}" == "${cur}"* ]]; then
					completions+=("${opt}")
			fi
	done
	COMPREPLY=("${completions[@]}")
	return 0
}
complete -F _%s_completion %s
`, rc.Name, strings.Join(rc.getCommandNames(), " "), rc.Name, rc.Name)
	return nil
}

// genZshCompletion generates the zsh completion script.
func (rc *RootCommand) genZshCompletion() error {
	fmt.Printf(`
#compdef %s

_%s_completion() {
        local line
        read -L line

        local commands="%s"
        local words=("${(s/ /)line}")
        local cur="${words[-1]}"

        if [[ $words[1] == "-*" ]]; then
                reply=("${(@(|)${(M)commands:#$cur} )}")
        else
                reply=("${(@(|)${(M)commands:#$cur} )}")
        fi
}

_%s_completion
`, rc.Name, rc.Name, strings.Join(rc.getCommandNames(), " "), rc.Name)
	return nil
}

// genFishCompletion generates the fish completion script.
func (rc *RootCommand) genFishCompletion() error {
	fmt.Printf(`
function __fish_%s_complete
        set -lx COMP_WORDS (commandline -o | string split -s ' ')
        set -lx COMP_CWORD (math (contains -i -- (commandline -o | string split -s ' ') -- (commandline -o | string cursor)) - 1)
        set -lx COMP_LINE (commandline -o)

        set -lx cmd (string split -s ' ' -- (commandline -o))[1]

        for word in (%s)
                if string match -q $word $COMP_WORDS[$COMP_CWORD]
                        echo $word
                end
        end
end

complete -f -c %s -a "(__fish_%s_complete)"
`, rc.Name, strings.Join(rc.getCommandNames(), " "), rc.Name, rc.Name)
	return nil
}
