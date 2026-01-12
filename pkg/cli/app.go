package cli

import (
	"context"
	"fmt"
	"maps"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/mirkobrombin/go-cli-builder/v2/pkg/help"
	"github.com/mirkobrombin/go-cli-builder/v2/pkg/log"
	"github.com/mirkobrombin/go-cli-builder/v2/pkg/parser"
	"github.com/mirkobrombin/go-cli-builder/v2/pkg/resolver"
	"github.com/mirkobrombin/go-struct-flags/v2/pkg/binder"
)

// applyBindings binds flags and args to the struct fields using the external binder library.
func applyBindings(node *parser.CommandNode, flags map[string]string, args []string, effectiveFlags map[string]*parser.FlagMetadata) error {
	val := node.Value
	if val.Kind() != reflect.Ptr && val.CanAddr() {
		val = val.Addr()
	}

	b, err := binder.NewBinder(val.Interface())
	if err != nil {
		return err
	}

	for name, meta := range effectiveFlags {
		m := meta

		switch m.Field.Kind() {
		case reflect.Bool:
			b.AddBool(name, func(v bool) error {
				m.Field.SetBool(v)
				return nil
			})
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if m.Field.Type() == reflect.TypeOf(time.Duration(0)) {
				b.AddDuration(name, func(v time.Duration) error {
					m.Field.SetInt(int64(v))
					return nil
				})
			} else {
				b.AddInt(name, func(v int64) error {
					m.Field.SetInt(v)
					return nil
				})
			}
		case reflect.String:
			b.AddStrings(name, func(v []string) error {
				if len(v) > 0 {
					m.Field.SetString(v[0])
				}
				return nil
			})
		default:
			if m.Field.Kind() == reflect.Slice && m.Field.Type().Elem().Kind() == reflect.String {
				b.AddStrings(name, func(v []string) error {
					s := reflect.MakeSlice(m.Field.Type(), len(v), len(v))
					for i, val := range v {
						s.Index(i).SetString(val)
					}
					m.Field.Set(s)
					return nil
				})
			}
		}
	}

	for name, meta := range effectiveFlags {
		var passedVal *string
		if v, ok := flags[name]; ok {
			passedVal = &v
		}

		valToBind, err := resolver.GetValue(passedVal, meta.Env, meta.Default, meta.Field.Kind() == reflect.Bool)
		if err != nil {
			return err
		}

		if meta.Required && valToBind == "" && meta.Field.Kind() != reflect.Bool {
			return fmt.Errorf("missing required flag: --%s", name)
		}

		if valToBind != "" {
			if err := b.Run(name, []string{valToBind}); err != nil {
				return fmt.Errorf("invalid value for flag --%s: %w", name, err)
			}
		}
	}

	return bindArgs(node, args)
}

// App represents a CLI application.
type App struct {
	RootNode   *parser.CommandNode
	Translator help.Translator
}

// New creates a new App from a root struct.
func New(root any) (*App, error) {
	rootNode, err := parser.Parse("root", root)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}
	return &App{RootNode: rootNode}, nil
}

// SetName sets the name of the root command.
func (a *App) SetName(name string) {
	if a.RootNode != nil {
		a.RootNode.Name = name
	}
}

// Reload re-parses the root struct to pick up dynamic changes (e.g. map entries).
//
// Example:
//
//	err := app.Reload()
func (a *App) Reload() error {
	if a.RootNode == nil {
		return nil
	}
	val := a.RootNode.Value
	if val.Kind() != reflect.Ptr && val.CanAddr() {
		val = val.Addr()
	}
	return parser.ParseStruct(a.RootNode, val)
}

// AddCommand adds a dynamic command to the application.
func (a *App) AddCommand(name string, cmd *parser.CommandNode) {
	if a.RootNode.Children == nil {
		a.RootNode.Children = make(map[string]*parser.CommandNode)
	}
	a.RootNode.Children[name] = cmd
}

// SetTranslator sets the translator for the application.
func (a *App) SetTranslator(tr help.Translator) {
	a.Translator = tr
}

// Run executes the application based on the provided root struct.
// It parses the CLI arguments, resolves commands, binds flags, infuses dependencies, and runs lifecycle hooks.
//
// Example:
//
//	func main() {
//		app := &CLI{}
//		if err := cli.Run(app); err != nil {
//			log.Fatal(err)
//		}
//	}
func Run(root any) error {
	app, err := New(root)
	if err != nil {
		return err
	}
	return app.Run()
}

// Run executes the application.
func (a *App) Run() error {
	args := os.Args[1:]

	targetNode, allFlags, err := resolveCommand(a.RootNode, args)
	if err != nil {
		fmt.Println(help.GenerateHelp(a.RootNode, a.Translator))
		return err
	}

	path := getPathToNode(a.RootNode, targetNode)
	effectiveFlags := make(map[string]*parser.FlagMetadata)

	for _, node := range path {
		maps.Copy(effectiveFlags, node.Flags)
	}

	for _, arg := range allFlags {
		if arg == "-h" || arg == "--help" {
			fmt.Print(help.GenerateHelp(targetNode, a.Translator))
			return nil
		}
	}

	parsedFlags, positionalArgs, err := parseArgs(allFlags, effectiveFlags)
	if err != nil {
		fmt.Printf("Error: %v\n\n", err)
		fmt.Print(help.GenerateHelp(targetNode, a.Translator))
		return err
	}

	if err := applyBindings(targetNode, parsedFlags, positionalArgs, effectiveFlags); err != nil {
		fmt.Printf("Error: %v\n\n", err)
		fmt.Print(help.GenerateHelp(targetNode, a.Translator))
		return err
	}

	for _, node := range path {
		injectDependencies(node)
	}

	for _, node := range path {
		if beforeRunner, ok := node.Value.Interface().(BeforeRunner); ok {
			if err := beforeRunner.Before(); err != nil {
				return err
			}
		} else if node.Value.CanAddr() {
			if beforeRunner, ok := node.Value.Addr().Interface().(BeforeRunner); ok {
				if err := beforeRunner.Before(); err != nil {
					return err
				}
			}
		}
	}

	executed := false
	if runner, ok := targetNode.Value.Interface().(Runner); ok {
		if err := runner.Run(); err != nil {
			return err
		}
		executed = true
	} else if targetNode.Value.CanAddr() {
		if runner, ok := targetNode.Value.Addr().Interface().(Runner); ok {
			if err := runner.Run(); err != nil {
				return err
			}
			executed = true
		}
	}

	if !executed {
		fmt.Print(help.GenerateHelp(targetNode, a.Translator))
	}

	for i := len(path) - 1; i >= 0; i-- {
		node := path[i]
		if afterRunner, ok := node.Value.Interface().(AfterRunner); ok {
			if err := afterRunner.After(); err != nil {
				return err
			}
		} else if node.Value.CanAddr() {
			if afterRunner, ok := node.Value.Addr().Interface().(AfterRunner); ok {
				if err := afterRunner.After(); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// resolveCommand traverses the tree, skipping flags to find subcommands.
func resolveCommand(root *parser.CommandNode, args []string) (*parser.CommandNode, []string, error) {
	current := root
	remaining := []string{}
	parsingCmds := true

	for i := range args {
		arg := args[i]

		if parsingCmds {
			if strings.HasPrefix(arg, "-") {
				remaining = append(remaining, arg)
			} else {
				if child, ok := current.Children[arg]; ok {
					current = child
				} else {
					// Not a subcommand, must be positional arg.
					parsingCmds = false
					remaining = append(remaining, arg)
				}
			}
		} else {
			remaining = append(remaining, arg)
		}
	}

	return current, remaining, nil
}

// getPathToNode reconstructs path from root to target (inefficient but safe).
func getPathToNode(root, target *parser.CommandNode) []*parser.CommandNode {
	if root == target {
		return []*parser.CommandNode{root}
	}
	for _, child := range root.Children {
		path := getPathToNode(child, target)
		if path != nil {
			return append([]*parser.CommandNode{root}, path...)
		}
	}
	return nil
}

// parseArgs parses flags based on effective metadata.
func parseArgs(args []string, effectiveFlags map[string]*parser.FlagMetadata) (map[string]string, []string, error) {
	flags := make(map[string]string)
	positionals := []string{}

	// Reverse Lookup for short flags
	shortMap := make(map[string]string)
	for name, meta := range effectiveFlags {
		if meta.Short != "" {
			shortMap[meta.Short] = name
		}
	}

	i := 0
	for i < len(args) {
		arg := args[i]

		if strings.HasPrefix(arg, "--") {
			name := arg[2:]
			value := ""
			hasValue := false

			if strings.Contains(name, "=") {
				parts := strings.SplitN(name, "=", 2)
				name = parts[0]
				value = parts[1]
				hasValue = true
			}

			meta, ok := effectiveFlags[name]
			if !ok {
				return nil, nil, fmt.Errorf("unknown flag: --%s", name)
			}

			if hasValue {
				flags[name] = value
			} else {
				if meta.Field.Kind() == reflect.Bool {
					flags[name] = "true"
				} else {
					if i+1 < len(args) {
						flags[name] = args[i+1]
						i++
					} else {
						return nil, nil, fmt.Errorf("flag needs an argument: --%s", name)
					}
				}
			}

		} else if strings.HasPrefix(arg, "-") {
			shorthand := arg[1:]
			name := shorthand
			value := ""
			hasValue := false

			if strings.Contains(name, "=") {
				parts := strings.SplitN(name, "=", 2)
				sName := parts[0]
				value = parts[1]
				hasValue = true

				if lName, ok := shortMap[sName]; ok {
					name = lName
				} else {
					return nil, nil, fmt.Errorf("unknown shorthand flag: -%s", sName)
				}
			} else {
				if lName, ok := shortMap[name]; ok {
					name = lName
				} else {
					if _, ok := effectiveFlags[name]; !ok {
						return nil, nil, fmt.Errorf("unknown shorthand flag: -%s", shorthand)
					}
				}
			}

			meta := effectiveFlags[name]

			if hasValue {
				flags[name] = value
			} else {
				if meta.Field.Kind() == reflect.Bool {
					flags[name] = "true"
				} else {
					if i+1 < len(args) {
						flags[name] = args[i+1]
						i++
					} else {
						return nil, nil, fmt.Errorf("flag needs an argument: -%s", shorthand)
					}
				}
			}

		} else {
			positionals = append(positionals, arg)
		}
		i++
	}

	return flags, positionals, nil
}

// bindArgs binds positional arguments to the struct fields using the internal resolver.
func bindArgs(node *parser.CommandNode, args []string) error {
	argIdx := 0
	for _, meta := range node.Args {
		if meta.IsGreedy {
			if len(args) > argIdx {
				for _, v := range args[argIdx:] {
					if err := resolver.BindValue(meta.Field, v); err != nil {
						return err
					}
				}
			} else if meta.Required {
				return fmt.Errorf("missing required positional arguments: %s", meta.Description)
			}
			break
		} else {
			if argIdx < len(args) {
				if err := resolver.BindValue(meta.Field, args[argIdx]); err != nil {
					return err
				}
				argIdx++
			} else if meta.Required {
				return fmt.Errorf("missing required argument: %s", meta.Description)
			}
		}
	}
	return nil
}

// injectDependencies injects the logger and context into the command struct if it embeds the Base struct.
func injectDependencies(node *parser.CommandNode) {
	logger := log.New()

	val := node.Value
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := val.Type().Field(i)

		if fieldType.Type == reflect.TypeFor[Base]() {
			if field.CanSet() {
				base := Base{
					Logger: logger,
					Ctx:    context.Background(),
				}
				field.Set(reflect.ValueOf(base))
			}
		}
	}
}
