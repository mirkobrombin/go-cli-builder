package parser

import (
	"reflect"
)

// FlagMetadata holds information about a flag.
type FlagMetadata struct {
	Name        string
	Short       string
	Description string
	Default     string
	Env         string
	Required    bool
	Field       reflect.Value
}

// ArgMetadata holds information about a positional argument.
type ArgMetadata struct {
	Description string
	Required    bool
	IsGreedy    bool
	Field       reflect.Value
}

// CommandNode represents a node in the command tree.
type CommandNode struct {
	Name        string
	Description string
	Aliases     []string
	Flags       map[string]*FlagMetadata
	ShortFlags  map[string]string // Maps short name to full name
	Args        []*ArgMetadata
	Children    map[string]*CommandNode
	Value       reflect.Value
	Type        reflect.Type
}

// NewCommandNode creates a new CommandNode with initialized maps.
//
// Example:
//
//	node := parser.NewCommandNode("init", "Initialize project", reflect.ValueOf(&InitCmd{}))
func NewCommandNode(name, description string, val reflect.Value) *CommandNode {
	return &CommandNode{
		Name:        name,
		Description: description,
		Flags:       make(map[string]*FlagMetadata),
		ShortFlags:  make(map[string]string),
		Children:    make(map[string]*CommandNode),
		Value:       val,
		Type:        val.Type(),
	}
}
