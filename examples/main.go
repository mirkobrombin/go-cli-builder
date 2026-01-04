package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mirkobrombin/go-cli-builder/v2/pkg/cli"
)

// CLI is the root application struct.
type CLI struct {
	Verbose bool `cli:"verbose,v" help:"Enable verbose output" env:"VERBOSE"`

	// Subcommands
	Add    AddCmd    `cmd:"" help:"Add a new item to the list"`
	Remove RemoveCmd `cmd:"" help:"Remove an item from the list"`
	List   ListCmd   `cmd:"" help:"List all items"`

	cli.Base
}

func (c *CLI) Before() error {
	if c.Verbose {
		c.Logger.Info("Verbose mode enabled")
	}
	return nil
}

// Storage logic
const dbFile = "items.json"

func loadItems() ([]string, error) {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return []string{}, nil
	}
	data, err := os.ReadFile(dbFile)
	if err != nil {
		return nil, err
	}
	var items []string
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func saveItems(items []string) error {
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(dbFile, data, 0644)
}

// AddCmd adds an item.
type AddCmd struct {
	Item string `arg:"" required:"true" help:"Item to add"`

	cli.Base
}

func (c *AddCmd) Run() error {
	items, err := loadItems()
	if err != nil {
		return err
	}
	items = append(items, c.Item)
	if err := saveItems(items); err != nil {
		return err
	}
	c.Logger.Success("Adding item: %s", c.Item)
	return nil
}

// RemoveCmd removes an item.
type RemoveCmd struct {
	Item string `arg:"" required:"true" help:"Item to remove"`

	cli.Base
}

func (c *RemoveCmd) Run() error {
	items, err := loadItems()
	if err != nil {
		return err
	}
	newItems := []string{}
	found := false
	for _, item := range items {
		if item == c.Item {
			found = true
			continue
		}
		newItems = append(newItems, item)
	}

	if !found {
		c.Logger.Warning("Item not found: %s", c.Item)
		return nil
	}

	if err := saveItems(newItems); err != nil {
		return err
	}
	c.Logger.Success("Removed item: %s", c.Item)
	return nil
}

// ListCmd lists all items.
type ListCmd struct {
	cli.Base
}

func (c *ListCmd) Run() error {
	items, err := loadItems()
	if err != nil {
		return err
	}
	c.Logger.Info("Listing items...")
	if len(items) == 0 {
		fmt.Println("(no items)")
		return nil
	}
	for _, item := range items {
		fmt.Printf("- %s\n", item)
	}
	return nil
}

func main() {
	app := &CLI{}
	if err := cli.Run(app); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
