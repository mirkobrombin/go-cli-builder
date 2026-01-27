package parser

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mirkobrombin/go-foundation/pkg/tags"
)

var flagTagParser = tags.NewParser("flag", tags.WithPairDelimiter(","), tags.WithKVSeparator(":"))

// Parse parses a struct and returns a CommandNode tree.
//
// Example:
//
//	root := &RootCmd{}
//	node, err := parser.Parse("apx", root)
func Parse(name string, root any) (*CommandNode, error) {
	val := reflect.ValueOf(root)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	node := NewCommandNode(name, "", val)

	if err := ParseStruct(node, val); err != nil {
		return nil, err
	}

	return node, nil
}

// ParseStruct recursively parses the struct fields to build the command tree.
//
// Example:
//
//	node := parser.NewCommandNode("init", "Initialize project", reflect.ValueOf(&InitCmd{}))
//	val := reflect.ValueOf(&InitCmd{})
//	err := parser.ParseStruct(node, val)
func ParseStruct(node *CommandNode, val reflect.Value) error {
	v := val
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	typ := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := v.Field(i)

		if field.Tag.Get("internal") == "ignore" {
			continue
		}

		cmdTag, ok := field.Tag.Lookup("cmd")

		// Pre-process for map merging
		if ok && cmdTag == "*" {
			fVal := fieldVal
			for fVal.Kind() == reflect.Ptr {
				if fVal.IsNil() {
					break
				}
				fVal = fVal.Elem()
			}

			if fVal.Kind() == reflect.Map {
				if !fVal.IsNil() {
					if node.Children == nil {
						node.Children = make(map[string]*CommandNode)
					}

					iter := fVal.MapRange()
					for iter.Next() {
						cmdName := iter.Key().String()
						val := iter.Value()

						// Unwrap pointer to reach the command struct
						startVal := val
						for startVal.Kind() == reflect.Ptr {
							if startVal.IsNil() {
								break
							}
							startVal = startVal.Elem()
						}

						if startVal.Kind() == reflect.Ptr && startVal.IsNil() {
							continue
						}

						helpPrefix := field.Tag.Get("help")
						description := ""
						if helpPrefix != "" {
							description = fmt.Sprintf("pr:%s.%s", helpPrefix, cmdName)
						}

						childNode := NewCommandNode(cmdName, description, startVal)
						node.Children[cmdName] = childNode

						if err := ParseStruct(childNode, startVal); err != nil {
							return err
						}
					}
				}
			}
			continue
		}

		if ok {
			startVal := fieldVal
			for startVal.Kind() == reflect.Ptr {
				if startVal.IsNil() {
					if startVal.CanSet() {
						startVal.Set(reflect.New(startVal.Type().Elem()))
					} else {
						break
					}
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

			description := field.Tag.Get("help")

			childNode := NewCommandNode(cmdName, description, startVal)
			childNode.Aliases = aliases

			node.Children[cmdName] = childNode
			for _, alias := range aliases {
				node.Children[alias] = childNode
			}

			if err := ParseStruct(childNode, startVal); err != nil {
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

		if flagTag, ok := field.Tag.Lookup("flag"); ok {
			meta := parseFlagTag(flagTag, fieldVal)

			name := meta.Name
			if name == "" {
				name = strings.ToLower(field.Name)
			}

			node.Flags[name] = meta
			if meta.Short != "" {
				node.ShortFlags[meta.Short] = name
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

// parseFlagTag parses the flag:"short:x, long:y, name:z" format.
func parseFlagTag(tag string, fieldVal reflect.Value) *FlagMetadata {
	meta := &FlagMetadata{Field: fieldVal}

	parsed := flagTagParser.Parse(tag)

	if vals, ok := parsed["short"]; ok && len(vals) > 0 {
		meta.Short = vals[0]
	}

	if vals, ok := parsed["long"]; ok && len(vals) > 0 {
		meta.Name = vals[0]
	}

	if vals, ok := parsed["name"]; ok && len(vals) > 0 {
		meta.Description = vals[0]
	}

	return meta
}
