# Go CLI Builder (V2)

A lightweight and flexible **declarative** library for building command-line interfaces (CLIs) 
in Go. This library provides a simple and intuitive way to define commands, 
flags (including short names), aliases and more using struct tags.

## ⚠️ Migration to V2

> V1 is now deprecated and not supported, please migrate to V2.

**Version 2.0 is a complete rewrite.** It moves from an imperative approach (calling methods to add flags) to a **declarative, code-first approach** (using struct tags). 
V1 code is **not compatible** with V2. Please refer to the [Basic Usage](#basic-usage) section to see the new pattern.

## Features

- **Declarative Command Definition:** Define commands and flags using struct tags (`cmd`, `cli`, `arg`, `help`).
- **Type-Safe Flag Handling:** Automatically binds flags to basic types (`int`, `bool`, `string`, `time.Duration`, `[]string`) and structs.
- **Dependency Injection:** Automatically injects `Logger` and `Context` into your commands via embedding.
- **Environment Variable Integration:** Map environment variables directly to flags using the `env:"VAR_NAME"` tag.
- **Built-in Help Generation:** Automatically generates formatted help messages based on your structs and tags.
- **Customizable Logging:** Includes a built-in logger (`Info`, `Success`, `Warning`, `Error`) available in every command.
- **Lifecycle Hooks:** Supports `Before()` and `After()` methods for command initialization and cleanup.

## Getting Started

### Installation

```bash
go get github.com/mirkobrombin/go-cli-builder/v2
```

### Basic Usage

```go
package main

import (
	"fmt"
	"os"

	"github.com/mirkobrombin/go-cli-builder/v2/pkg/cli"
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

For more detailed examples, check the `examples/v2` directory.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file 
for details.