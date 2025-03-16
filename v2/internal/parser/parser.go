package parser

import (
	"reflect"
	"strings"
)

// Parse parses a struct and returns a CommandNode tree.
//
// Example:
//
//	type CLI struct {
//		Serve ServeCmd `cmd:"serve"`
//	}
//	node, err := parser.Parse(&CLI{})
func Parse(root any) (*CommandNode, error) {
	val := reflect.ValueOf(root)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	node := NewCommandNode("root", "", val)

	if err := parseStruct(node, val); err != nil {
		return nil, err
	}

	return node, nil
}

// parseStruct recursively parses the struct fields to build the command tree.
func parseStruct(node *CommandNode, val reflect.Value) error {
	typ := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		if field.Tag.Get("internal") == "ignore" {
			continue
		}

		if cmdTag, ok := field.Tag.Lookup("cmd"); ok {
			startVal := fieldVal
			if startVal.Kind() == reflect.Ptr {
				if startVal.IsNil() {
					startVal.Set(reflect.New(startVal.Type().Elem()))
				}
				startVal = startVal.Elem()
			}

			cmdName := strings.ToLower(field.Name)
			if cmdTag != "" {
				cmdName = cmdTag
			}

			var aliases []string
			if aliasTag, ok := field.Tag.Lookup("aliases"); ok {
				aliases = strings.Split(aliasTag, ",")
			}

			help := field.Tag.Get("help")

			childNode := NewCommandNode(cmdName, help, startVal)
			childNode.Aliases = aliases

			node.Children[cmdName] = childNode
			for _, alias := range aliases {
				node.Children[alias] = childNode
			}

			if err := parseStruct(childNode, startVal); err != nil {
				return err
			}
			continue
		}

		if cliTag, ok := field.Tag.Lookup("cli"); ok {
			parts := strings.Split(cliTag, ",")
			name := parts[0]
			short := ""
			if len(parts) > 1 {
				short = parts[1]
			}

			required := false
			if reqTag, ok := field.Tag.Lookup("required"); ok && reqTag == "true" {
				required = true
			}

			flagMeta := &FlagMetadata{
				Name:        name,
				Short:       short,
				Description: field.Tag.Get("help"),
				Default:     field.Tag.Get("default"),
				Env:         field.Tag.Get("env"),
				Required:    required,
				Field:       fieldVal,
			}

			node.Flags[name] = flagMeta
			if short != "" {
				node.ShortFlags[short] = name
			}
			continue
		}

		if _, ok := field.Tag.Lookup("arg"); ok {
			required := false
			if reqTag, ok := field.Tag.Lookup("required"); ok && reqTag == "true" {
				required = true
			}

			isGreedy := field.Type.Kind() == reflect.Slice && field.Type.Elem().Kind() == reflect.String

			argMeta := &ArgMetadata{
				Description: field.Tag.Get("help"),
				Required:    required,
				IsGreedy:    isGreedy,
				Field:       fieldVal,
			}

			node.Args = append(node.Args, argMeta)
			continue
		}
	}
	return nil
}
