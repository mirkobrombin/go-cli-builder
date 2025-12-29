# Command Management

In V2, commands are defined as **Go structs**. The library uses reflection and struct tags to build the CLI tree.

## Defining a Command

A command is simply a struct that:
1. (Optional) Embeds `cli.Base` to get access to `Logger` and `Context`.
2. (Optional) Implements the `Runner` interface (`Run() error`).
3. (Optional) Implements lifecycle hooks (`Before() error`, `After() error`).
4. Defines fields for flags and arguments.

```go
type MyCommand struct {
    cli.Base
    
    // Flags
    Verbose bool `cli:"verbose,v" help:"Enable verbose mode"`
    
    // Positional Arguments
    Files []string `arg:"" help:"Files to process"`
}

func (c *MyCommand) Run() error {
    c.Logger.Info("Processing files: %v", c.Files)
    return nil
}
```

## Subcommands

Subcommands are defined as fields within a command struct, tagged with `cmd:""`.

```go
type RootCLI struct {
    // Defines a subcommand "start"
    Start StartCmd `cmd:"start" help:"Start the server"`
}
```

The field name is used as the command name (lowercased) unless explicitly overridden in the tag (e.g., `cmd:"my-start"`).

## Lifecycle Hooks

You can define logic to run before or after a command execution:

```go
func (c *MyCommand) Before() error {
    // Validation or setup logic
    return nil
}

func (c *MyCommand) After() error {
    // Cleanup logic
    return nil
}
```