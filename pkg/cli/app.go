package cli

import (
	"context"
	"fmt"
	"maps"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/mirkobrombin/go-cli-builder/v3/internal/binder"
	"github.com/mirkobrombin/go-cli-builder/v3/pkg/help"
	"github.com/mirkobrombin/go-cli-builder/v3/pkg/log"
	"github.com/mirkobrombin/go-cli-builder/v3/pkg/parser"
	"github.com/mirkobrombin/go-cli-builder/v3/pkg/resolver"
	"github.com/mirkobrombin/go-foundation/v2/core/errutil"
	"github.com/mirkobrombin/go-foundation/v2/core/reflectutil"
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
	RootNode      *parser.CommandNode
	Translator    help.Translator
	Container     Container
	Validator     Validator
	ctx           context.Context
	logger        log.Logger
	version       string
	mu            sync.Mutex
	panicRecovery bool
}

// AppOption is a functional option for App configuration.
type AppOption func(*App)

// WithVersion sets the application version.
func WithVersion(v string) AppOption {
	return func(a *App) { a.version = v }
}

// WithContext sets the application context.
func WithContext(ctx context.Context) AppOption {
	return func(a *App) { a.ctx = ctx }
}

// WithTranslator sets the help translator.
func WithTranslator(tr help.Translator) AppOption {
	return func(a *App) { a.Translator = tr }
}

// WithContainer sets the DI container for dependency injection in commands.
//
// Any value satisfying Container is accepted, including *di.Container from
// go-foundation/v2/app/di.
func WithContainer(container Container) AppOption {
	return func(a *App) { a.Container = container }
}

// WithValidator sets the struct validator for command validation.
//
// Use ValidatorFunc to adapt a validator with a different signature.
func WithValidator(validator Validator) AppOption {
	return func(a *App) { a.Validator = validator }
}

// WithLogger sets the logger handed to every command through Base.
//
// Without it commands get log.New(), which writes readable lines to the
// terminal. Pass fdn.NewStructured to emit JSON instead, or fdn.Wrap to reuse
// a logger the application already built.
func WithLogger(logger log.Logger) AppOption {
	return func(a *App) { a.logger = logger }
}

// WithPanicRecovery enables panic recovery in Run.
func WithPanicRecovery() AppOption {
	return func(a *App) { a.panicRecovery = true }
}

// New creates a new App from a root struct.
//
// Example:
//
//	app, err := cli.New(&CLI{}, cli.WithVersion("1.0.0"))
func New(root any, opts ...AppOption) (*App, error) {
	rootNode, err := parser.Parse("root", root)
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
	app := &App{RootNode: rootNode, ctx: ctx}
	for _, opt := range opts {
		if opt != nil {
			opt(app)
		}
	}
	if app.logger == nil {
		app.logger = log.New()
	}
	return app, nil
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
	a.mu.Lock()
	defer a.mu.Unlock()
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
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.RootNode.Children == nil {
		a.RootNode.Children = make(map[string]*parser.CommandNode)
	}
	a.RootNode.Children[name] = cmd
}

// SetTranslator sets the translator for the application.
func (a *App) SetTranslator(tr help.Translator) {
	a.Translator = tr
}

// SetVersion sets the version string for the application.
func (a *App) SetVersion(v string) {
	a.version = v
}

// SetContext sets the context for the application.
func (a *App) SetContext(ctx context.Context) {
	a.ctx = ctx
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
func Run(root any, opts ...AppOption) error {
	app, err := New(root, opts...)
	if err != nil {
		return err
	}
	return app.Run()
}

// Run executes the application.
func (a *App) Run() (runErr error) {
	if a.panicRecovery {
		defer func() {
			if r := recover(); r != nil {
				runErr = errutil.Wrap(fmt.Errorf("panic: %v", r))
			}
		}()
	}

	args := os.Args[1:]

	targetNode, allFlags, err := resolveCommand(a.RootNode, args)
	if err != nil {
		fmt.Println(help.GenerateHelp(a.RootNode, a.Translator))
		return errutil.Wrap(err)
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

	for i, arg := range allFlags {
		if arg == "--version" && a.version != "" {
			fmt.Println(a.version)
			return nil
		}
		if arg == "--completion" && i+1 < len(allFlags) {
			shell := allFlags[i+1]
			switch shell {
			case "bash":
				return a.GenBashCompletion(os.Stdout)
			case "zsh":
				return a.GenZshCompletion(os.Stdout)
			case "fish":
				return a.GenFishCompletion(os.Stdout)
			default:
				return errutil.Wrap(fmt.Errorf("unknown shell for completion: %s (supported: bash, zsh, fish)", shell))
			}
		}
	}

	parsedFlags, positionalArgs, err := parseArgs(allFlags, effectiveFlags)
	if err != nil {
		fmt.Printf("Error: %v\n\n", err)
		fmt.Print(help.GenerateHelp(targetNode, a.Translator))
		return errutil.Wrap(err)
	}

	if err := applyBindings(targetNode, parsedFlags, positionalArgs, effectiveFlags); err != nil {
		fmt.Printf("Error: %v\n\n", err)
		fmt.Print(help.GenerateHelp(targetNode, a.Translator))
		return errutil.Wrap(err)
	}

	if a.Validator != nil {
		val := targetNode.Value
		if val.Kind() != reflect.Ptr && val.CanAddr() {
			val = val.Addr()
		}
		if err := a.Validator.Validate(val.Interface()); err != nil {
			fmt.Fprintf(os.Stderr, "validation error: %s\n", err.Error())
			fmt.Print(help.GenerateHelp(targetNode, a.Translator))
			return errutil.Wrap(err)
		}
	}

	for _, node := range path {
		if err := injectDependencies(node, a.ctx, a.logger, a.Container); err != nil {
			return errutil.Wrap(err)
		}
	}

	for _, node := range path {
		if beforeRunner, ok := node.Value.Interface().(BeforeRunner); ok {
			if err := beforeRunner.Before(); err != nil {
				return errutil.Wrap(err)
			}
		} else if node.Value.CanAddr() {
			if beforeRunner, ok := node.Value.Addr().Interface().(BeforeRunner); ok {
				if err := beforeRunner.Before(); err != nil {
					return errutil.Wrap(err)
				}
			}
		}
	}

	executed := false
	if runner, ok := targetNode.Value.Interface().(Runner); ok {
		if err := runner.Run(); err != nil {
			return errutil.Wrap(err)
		}
		executed = true
	} else if targetNode.Value.CanAddr() {
		if runner, ok := targetNode.Value.Addr().Interface().(Runner); ok {
			if err := runner.Run(); err != nil {
				return errutil.Wrap(err)
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
				return errutil.Wrap(err)
			}
		} else if node.Value.CanAddr() {
			if afterRunner, ok := node.Value.Addr().Interface().(AfterRunner); ok {
				if err := afterRunner.After(); err != nil {
					return errutil.Wrap(err)
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
					// A command group (has subcommands, takes no positional
					// args of its own) cannot accept this token: it is an
					// unknown command, not a positional argument.
					if len(current.Children) > 0 && len(current.Args) == 0 {
						return current, remaining, fmt.Errorf("unknown command: %q", arg)
					}
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

// getPathToNode reconstructs path from root to target.
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

// bindArgs binds positional arguments to the struct fields using reflectutil.
func bindArgs(node *parser.CommandNode, args []string) error {
	argIdx := 0
	for _, meta := range node.Args {
		if meta.IsGreedy {
			if len(args) > argIdx {
				for _, v := range args[argIdx:] {
					if err := reflectutil.Bind(meta.Field, v); err != nil {
						return err
					}
				}
			} else if meta.Required {
				return fmt.Errorf("missing required positional arguments: %s", meta.Description)
			}
			break
		} else {
			if argIdx < len(args) {
				if err := reflectutil.Bind(meta.Field, args[argIdx]); err != nil {
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

// injectDependencies injects the logger, context and DI container into the command struct if it embeds the Base struct.
func injectDependencies(node *parser.CommandNode, ctx context.Context, logger log.Logger, container Container) error {
	if logger == nil {
		logger = log.New()
	}

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
					Logger:    logger,
					Ctx:       ctx,
					Container: container,
				}
				field.Set(reflect.ValueOf(base))
			}
		}
	}

	if container != nil {
		return injectDIFields(val, container)
	}

	return nil
}

// injectDIFields uses the DI container to inject dependencies tagged with `inject`.
//
// A missing or mistyped dependency is reported instead of being swallowed: in
// Foundation v2 the container answers with an error rather than leaving the
// field at its zero value.
func injectDIFields(val reflect.Value, container Container) error {
	if val.Kind() != reflect.Struct || !val.CanAddr() {
		return nil
	}

	if err := container.Inject(val.Addr().Interface()); err != nil {
		return fmt.Errorf("dependency injection failed: %w", err)
	}

	return nil
}
