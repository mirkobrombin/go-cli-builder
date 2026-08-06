# Go CLI Builder (V3)

A lightweight and flexible **declarative** library for building command-line interfaces (CLIs) 
in Go. This library provides a simple and intuitive way to define commands, 
flags (including short names), aliases and more using struct tags.

## Migration to V3

Update the import path, then apply the table below. `WithValidator` is the only
call that needs more than a rewrite.

| V2 | V3 |
|---|---|
| `import ".../go-cli-builder/v2/..."` | `import ".../go-cli-builder/v3/..."` |
| `WithContainer(*di.Container)` | `WithContainer(cli.Container)`, satisfied by `*di.Container` as it is |
| `WithValidator(*validation.Validator)` | `WithValidator(cli.Validator)`, use `cli.ValidatorFunc` to adapt |
| `Base.Container` is `*di.Container` | `Base.Container` is `cli.Container` |
| `AppOption = options.Option[App]` | `AppOption func(*App)` |
| `log.NewFoundationLogger()` | removed, use `cli.WithLogger` with `pkg/log/fdn` |

The types above are declared by this library instead of imported from
`go-foundation`, so a future major of that library is no longer a breaking
change here.

### Behaviour changes

- **Command output is readable.** From v2.1.0 the logger emitted one JSON object
  per line, so `Logger.Info` printed `{"level":"info",...}` to the user and
  `Logger.Success` was indistinguishable from `Logger.Info`. V3 prints readable,
  colored lines. Pass `fdn.NewStructured` if a machine reads the output.
- **Warnings and errors go to stderr.** Informational and success messages stay
  on stdout, so piping a command's output no longer swallows its errors.

### Adapting the Foundation validator

```go
v := validation.New()
cli.Run(app, cli.WithValidator(cli.ValidatorFunc(func(target any) error {
	if errs := v.Validate(target); len(errs) > 0 {
		return errs
	}
	return nil
})))
```

## Features

- **Declarative Command Definition:** Define commands and flags using struct tags (`cmd`, `cli`, `arg`, `help`).
- **Type-Safe Flag Handling:** Automatically binds flags to basic types (`int`, `bool`, `string`, `time.Duration`, `[]string`) and structs.
- **Dependency Injection:** Automatically injects `Logger`, `Context` and an optional DI container into your commands via embedding.
- **Environment Variable Integration:** Map environment variables directly to flags using the `env:"VAR_NAME"` tag.
- **Built-in Help Generation:** Automatically generates formatted help messages based on your structs and tags.
- **Customizable Logging:** A built-in readable logger (`Info`, `Success`, `Warning`, `Error`) in every command, swappable with `WithLogger` for structured output.
- **Contained Dependencies:** No third party type appears in the public API, so upgrading what runs underneath is not a breaking change for you.
- **Lifecycle Hooks:** Supports `Before()` and `After()` methods for command initialization and cleanup.

## Getting Started

### Installation

```bash
go get github.com/mirkobrombin/go-cli-builder/v3
```

### Basic Usage

```go
package main

import (
	"fmt"
	"os"

	"github.com/mirkobrombin/go-cli-builder/v3/pkg/cli"
)

// Define your root CLI struct
type CLI struct {
	// Global flags
	Verbose bool `cli:"verbose,v" help:"Enable verbose output" env:"VERBOSE"`

	// Subcommands
	Add  AddCmd  `cmd:"add" help:"Add a new item"`
	List ListCmd `cmd:"list" help:"List all items"`

	// Embed Base to get Logger and Context
	cli.Base
}

// Optional: Lifecycle hook
func (c *CLI) Before() error {
	if c.Verbose {
		c.Logger.Info("Verbose mode enabled")
	}
	return nil
}

type AddCmd struct {
	Item string `arg:"" required:"true" help:"Item to add"`
	cli.Base
}

// Run is the entry point for the command
func (c *AddCmd) Run() error {
	c.Logger.Success("Adding item: %s", c.Item)
	return nil
}

type ListCmd struct {
	cli.Base
}

func (c *ListCmd) Run() error {
	c.Logger.Info("Listing items...")
	return nil
}

func main() {
	app := &CLI{}

	// Run the app - the library handles parsing, binding, and execution
	if err := cli.Run(app); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
```

## Documentation

For more detailed examples, check the `examples` directory.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file 
for details.