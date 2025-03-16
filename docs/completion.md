# Shell Completion

The Go CLI Builder library includes a built-in `completion` command that can 
generate shell completion scripts for Bash, Zsh, and Fish. This allows users 
to easily auto-complete commands and flags as they type.

## Usage

To generate a completion script, run the `completion` command followed by the 
name of your shell:

```bash
# For Bash:
./mycli completion bash > /etc/bash_completion.d/mycli
# Or, for Zsh:
./mycli completion zsh > ~/.zsh/completion/_mycli
# Or, for Fish:
./mycli completion fish > ~/.config/fish/completions/mycli.fish
```

**Note:** The exact location where you need to save the completion script 
might vary depending on your operating system and shell configuration. You 
might also need to source the completion script in your shell configuration 
file (e.g., `.bashrc`, `.zshrc`, `.config/fish/config.fish`).

## How it Works

When you run the `completion` command, the library generates a shell-specific 
script that defines how your CLI's commands and subcommands should be 
auto-completed. This script is then used by your shell to provide suggestions 
as you type commands.

The completion script automatically includes the names of all your commands 
and subcommands. For more advanced completion features (e.g., completing flag 
values), you might need to customize the generated script.