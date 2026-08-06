package cli

import (
	"context"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/mirkobrombin/go-cli-builder/v3/pkg/parser"
	"github.com/mirkobrombin/go-foundation/v2/app/di"
	"github.com/mirkobrombin/go-foundation/v2/core/validation"
)

// foundationValidator adapts the Foundation validator to the Validator
// interface, the way a consumer is expected to.
func foundationValidator(v *validation.Validator) Validator {
	return ValidatorFunc(func(target any) error {
		if errs := v.Validate(target); len(errs) > 0 {
			return errs
		}
		return nil
	})
}

// The Foundation container must satisfy the Container interface without an
// adapter. This assertion fails at compile time if that stops being true.
var _ Container = (*di.Container)(nil)

var (
	errBeforeFailed = errors.New("before failed")
	errAfterFailed  = errors.New("after failed")
)

type testRootCmd struct {
	Verbose bool   `cli:"verbose,v" help:"Enable verbose"`
	Name    string `cli:"name" help:"Your name" default:"world"`

	Add    testAddCmd    `cmd:"" help:"Add an item"`
	Remove testRemoveCmd `cmd:"" help:"Remove an item"`
}

func (c *testRootCmd) Run() error { return nil }

type testAddCmd struct {
	Item string `arg:"" required:"true" help:"Item to add"`
}

func (c *testAddCmd) Run() error { return nil }

type testRemoveCmd struct {
	Item string `arg:"" required:"true" help:"Item to remove"`
}

func (c *testRemoveCmd) Run() error { return nil }

type testSimpleCmd struct {
	Value int `cli:"value" help:"A value"`
}

func makeNode(name, desc string, v any, opts ...func(*parser.CommandNode)) *parser.CommandNode {
	rv := reflect.ValueOf(v)
	rv2 := rv
	for rv2.Kind() == reflect.Ptr {
		rv2 = rv2.Elem()
	}
	node := parser.NewCommandNode(name, desc, rv)
	for _, o := range opts {
		o(node)
	}
	return node
}

func TestNew_ValidRoot(t *testing.T) {
	app, err := New(&testRootCmd{})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	if app.RootNode == nil {
		t.Fatal("expected non-nil RootNode")
	}
	if app.RootNode.Name != "root" {
		t.Errorf("got name %q, want %q", app.RootNode.Name, "root")
	}
	if len(app.RootNode.Children) != 2 {
		t.Errorf("expected 2 children, got %d", len(app.RootNode.Children))
	}
}

func TestNew_InvalidRoot(t *testing.T) {
	app, err := New(42)
	if err != nil {
		t.Fatalf("New should not error even for non-struct: %v", err)
	}
	if app.RootNode == nil {
		t.Fatal("expected non-nil RootNode")
	}
}

func TestSetName(t *testing.T) {
	app, err := New(&testRootCmd{})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	app.SetName("myapp")
	if app.RootNode.Name != "myapp" {
		t.Errorf("got name %q, want %q", app.RootNode.Name, "myapp")
	}
}

func TestAddCommand(t *testing.T) {
	app, err := New(&testRootCmd{})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	child := makeNode("list", "List items", &testSimpleCmd{})
	app.AddCommand("list", child)
	if _, ok := app.RootNode.Children["list"]; !ok {
		t.Fatal("expected 'list' in children")
	}
}

func TestResolveCommand_Direct(t *testing.T) {
	app, _ := New(&testRootCmd{})
	args := []string{"add", "myitem"}
	node, remaining, err := resolveCommand(app.RootNode, args)
	if err != nil {
		t.Fatalf("resolveCommand failed: %v", err)
	}
	if node.Name != "add" {
		t.Errorf("got node %q, want %q", node.Name, "add")
	}
	if len(remaining) != 1 || remaining[0] != "myitem" {
		t.Errorf("got remaining %v, want [myitem]", remaining)
	}
}

func TestResolveCommand_Nested(t *testing.T) {
	type nested struct {
		Inner struct {
			AddCmd testAddCmd `cmd:"add" help:"Add command"`
		} `cmd:"inner" help:"Inner commands"`
	}
	app, _ := New(&nested{})
	args := []string{"inner", "add", "val"}
	node, _, err := resolveCommand(app.RootNode, args)
	if err != nil {
		t.Fatalf("resolveCommand failed: %v", err)
	}
	if node.Name != "add" {
		t.Errorf("got node %q, want %q", node.Name, "add")
	}
}

func TestResolveCommand_FlagsFirst(t *testing.T) {
	app, _ := New(&testRootCmd{})
	args := []string{"--verbose", "add", "myitem"}
	node, remaining, err := resolveCommand(app.RootNode, args)
	if err != nil {
		t.Fatalf("resolveCommand failed: %v", err)
	}
	if node.Name != "add" {
		t.Errorf("got node %q, want %q", node.Name, "add")
	}
	if len(remaining) < 2 {
		t.Errorf("expected at least 2 remaining (flag+arg), got %v", remaining)
	}
}

func TestResolveCommand_UnknownCommand(t *testing.T) {
	app, _ := New(&testRootCmd{})
	args := []string{"unknown"}
	// A command group must reject an unknown subcommand rather than swallowing
	// it as a positional and falling back to root help.
	if _, _, err := resolveCommand(app.RootNode, args); err == nil {
		t.Fatal("expected error for unknown command, got nil")
	}
}

// flagValue returns the one value parsed for name, and fails when the flag was
// seen any other number of times.
func flagValue(t *testing.T, parsed map[string][]string, name string) string {
	t.Helper()
	v := parsed[name]
	if len(v) != 1 {
		t.Fatalf("flag %s: got %d values %v, want 1", name, len(v), v)
	}
	return v[0]
}

func TestParseArgs_LongFlag(t *testing.T) {
	flags := map[string]*parser.FlagMetadata{
		"name": {
			Name:  "name",
			Field: reflect.ValueOf(""),
		},
	}
	parsed, pos, err := parseArgs([]string{"--name", "hello"}, flags)
	if err != nil {
		t.Fatalf("parseArgs failed: %v", err)
	}
	if flagValue(t, parsed, "name") != "hello" {
		t.Errorf("got %v, want %q", parsed["name"], "hello")
	}
	if len(pos) != 0 {
		t.Errorf("expected no positional args, got %v", pos)
	}
}

func TestParseArgs_LongFlagWithEquals(t *testing.T) {
	flags := map[string]*parser.FlagMetadata{
		"name": {
			Name:  "name",
			Field: reflect.ValueOf(""),
		},
	}
	parsed, _, err := parseArgs([]string{"--name=hello"}, flags)
	if err != nil {
		t.Fatalf("parseArgs failed: %v", err)
	}
	if flagValue(t, parsed, "name") != "hello" {
		t.Errorf("got %v, want %q", parsed["name"], "hello")
	}
}

func TestParseArgs_ShortFlag(t *testing.T) {
	flags := map[string]*parser.FlagMetadata{
		"verbose": {
			Name:  "verbose",
			Short: "v",
			Field: reflect.ValueOf(true),
		},
	}
	node := makeNode("root", "", &testSimpleCmd{})
	node.Flags = flags
	node.ShortFlags["v"] = "verbose"

	parsed, _, err := parseArgs([]string{"-v", "true"}, flags)
	if err != nil {
		t.Fatalf("parseArgs failed: %v", err)
	}
	if flagValue(t, parsed, "verbose") != "true" {
		t.Errorf("got %v, want %q", parsed["verbose"], "true")
	}
}

func TestParseArgs_BoolFlag(t *testing.T) {
	flags := map[string]*parser.FlagMetadata{
		"verbose": {
			Name:  "verbose",
			Field: reflect.ValueOf(true),
		},
	}
	parsed, _, err := parseArgs([]string{"--verbose"}, flags)
	if err != nil {
		t.Fatalf("parseArgs failed: %v", err)
	}
	if flagValue(t, parsed, "verbose") != "true" {
		t.Errorf("got %q, want 'true' for bool flag", flagValue(t, parsed, "verbose"))
	}
}

func TestParseArgs_UnknownFlag(t *testing.T) {
	flags := map[string]*parser.FlagMetadata{}
	_, _, err := parseArgs([]string{"--unknown"}, flags)
	if err == nil {
		t.Error("expected error for unknown flag")
	}
}

func TestParseArgs_MissingValue(t *testing.T) {
	flags := map[string]*parser.FlagMetadata{
		"name": {
			Name:  "name",
			Field: reflect.ValueOf(""),
		},
	}
	_, _, err := parseArgs([]string{"--name"}, flags)
	if err == nil {
		t.Error("expected error for missing flag value")
	}
}

func TestParseArgs_PositionalArgs(t *testing.T) {
	flags := map[string]*parser.FlagMetadata{
		"verbose": {
			Name:  "verbose",
			Field: reflect.ValueOf(true),
		},
	}
	parsed, pos, err := parseArgs([]string{"--verbose", "file1", "file2"}, flags)
	if err != nil {
		t.Fatalf("parseArgs failed: %v", err)
	}
	if flagValue(t, parsed, "verbose") != "true" {
		t.Errorf("expected verbose=true, got %v", flagValue(t, parsed, "verbose"))
	}
	if len(pos) != 2 || pos[0] != "file1" || pos[1] != "file2" {
		t.Errorf("expected [file1 file2], got %v", pos)
	}
}

func TestBindArgs_Required(t *testing.T) {
	var val string
	node := parser.NewCommandNode("add", "", reflect.ValueOf(&struct {
		Item string
	}{}))
	node.Args = append(node.Args, &parser.ArgMetadata{
		Description: "item",
		Required:    true,
		Field:       reflect.ValueOf(&val).Elem(),
	})

	if err := bindArgs(node, []string{"myitem"}); err != nil {
		t.Fatalf("bindArgs failed: %v", err)
	}
	if val != "myitem" {
		t.Errorf("got %q, want %q", val, "myitem")
	}
}

func TestBindArgs_MissingRequired(t *testing.T) {
	var val string
	node := parser.NewCommandNode("add", "", reflect.ValueOf(&struct {
		Item string
	}{}))
	node.Args = append(node.Args, &parser.ArgMetadata{
		Description: "item",
		Required:    true,
		Field:       reflect.ValueOf(&val).Elem(),
	})

	if err := bindArgs(node, []string{}); err == nil {
		t.Error("expected error for missing required arg")
	}
}

func TestBindArgs_GreedySlice(t *testing.T) {
	var val []string
	node := parser.NewCommandNode("add", "", reflect.ValueOf(&struct {
		Items []string
	}{}))
	sliceField := reflect.ValueOf(&val).Elem()
	node.Args = append(node.Args, &parser.ArgMetadata{
		Description: "files",
		Required:    true,
		IsGreedy:    true,
		Field:       sliceField,
	})

	if err := bindArgs(node, []string{"a", "b", "c"}); err != nil {
		t.Fatalf("bindArgs failed: %v", err)
	}
	if len(val) != 3 || val[0] != "a" || val[1] != "b" || val[2] != "c" {
		t.Errorf("got %v, want [a b c]", val)
	}
}

func TestBindArgs_MultiplePositional(t *testing.T) {
	type args struct {
		From string
		To   string
	}
	a := args{}
	node := parser.NewCommandNode("copy", "", reflect.ValueOf(&a))
	node.Args = append(node.Args,
		&parser.ArgMetadata{Description: "from", Required: true, Field: reflect.ValueOf(&a.From).Elem()},
		&parser.ArgMetadata{Description: "to", Required: true, Field: reflect.ValueOf(&a.To).Elem()},
	)

	if err := bindArgs(node, []string{"src", "dst"}); err != nil {
		t.Fatalf("bindArgs failed: %v", err)
	}
	if a.From != "src" || a.To != "dst" {
		t.Errorf("got %+v, want {From:src To:dst}", a)
	}
}

func TestReload(t *testing.T) {
	app, _ := New(&testRootCmd{})
	if err := app.Reload(); err != nil {
		t.Fatalf("Reload failed: %v", err)
	}
	if app.RootNode == nil {
		t.Fatal("RootNode should not be nil after reload")
	}
}

func TestGetPathToNode(t *testing.T) {
	app, _ := New(&testRootCmd{})
	addNode := app.RootNode.Children["add"]
	if addNode == nil {
		t.Fatal("expected add child node")
	}
	path := getPathToNode(app.RootNode, addNode)
	if len(path) != 2 {
		t.Fatalf("expected path length 2, got %d", len(path))
	}
	if path[0] != app.RootNode {
		t.Error("path[0] should be root")
	}
	if path[1] != addNode {
		t.Error("path[1] should be add")
	}
}

func TestGetPathToNode_SameNode(t *testing.T) {
	node := makeNode("root", "", &testRootCmd{})
	path := getPathToNode(node, node)
	if len(path) != 1 || path[0] != node {
		t.Error("path to self should be just [self]")
	}
}

func TestInjectDependencies(t *testing.T) {
	type cmdWithBase struct {
		Base
	}
	v := &cmdWithBase{}
	ctx := context.Background()
	node := parser.NewCommandNode("test", "", reflect.ValueOf(v))
	if err := injectDependencies(node, ctx, nil, nil); err != nil {
		t.Fatalf("injectDependencies failed: %v", err)
	}
	if v.Ctx == nil {
		t.Error("expected non-nil Ctx after injectDependencies")
	}
	if v.Logger == nil {
		t.Error("expected non-nil Logger after injectDependencies")
	}
}

func TestInjectDependencies_NoBase(t *testing.T) {
	type cmdWithoutBase struct {
		Name string
	}
	v := &cmdWithoutBase{Name: "test"}
	ctx := context.Background()
	node := parser.NewCommandNode("test", "", reflect.ValueOf(v))
	originalName := v.Name
	if err := injectDependencies(node, ctx, nil, nil); err != nil {
		t.Fatalf("injectDependencies failed: %v", err)
	}
	if v.Name != originalName {
		t.Errorf("Name should not change: got %q, want %q", v.Name, originalName)
	}
}

func TestRunFunction(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"app", "add", "testitem"}
	err := Run(&testRootCmd{})
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}
}

type lifecycleCmd struct {
	Ran       bool
	BeforeRan bool
	AfterRan  bool

	Sub lifecycleSubCmd `cmd:"" help:"sub"`
}

func (c *lifecycleCmd) Before() error {
	c.BeforeRan = true
	return nil
}

func (c *lifecycleCmd) Run() error {
	c.Ran = true
	return nil
}

func (c *lifecycleCmd) After() error {
	c.AfterRan = true
	return nil
}

type lifecycleSubCmd struct {
	Ran       bool
	BeforeRan bool
	AfterRan  bool
}

func (c *lifecycleSubCmd) Before() error {
	c.BeforeRan = true
	return nil
}

func (c *lifecycleSubCmd) Run() error {
	c.Ran = true
	return nil
}

func (c *lifecycleSubCmd) After() error {
	c.AfterRan = true
	return nil
}

func TestRun_LifecycleOrder(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	cmd := &lifecycleCmd{}
	os.Args = []string{"app", "sub"}
	if err := Run(cmd); err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !cmd.BeforeRan {
		t.Error("expected root Before to be called")
	}
	if !cmd.Sub.BeforeRan {
		t.Error("expected sub Before to be called")
	}
	if !cmd.Sub.Ran {
		t.Error("expected sub Run to be called")
	}
	if !cmd.Sub.AfterRan {
		t.Error("expected sub After to be called")
	}
	if !cmd.AfterRan {
		t.Error("expected root After to be called")
	}
}

type beforeErrorCmd struct {
	ShouldFail bool
}

func (c *beforeErrorCmd) Before() error {
	if c.ShouldFail {
		return errBeforeFailed
	}
	return nil
}

func (c *beforeErrorCmd) Run() error { return nil }

type runWithHelpCmd struct {
	Name string `cli:"name" help:"Your name"`
}

func (c *runWithHelpCmd) Run() error { return nil }

func TestRun_HelpFlag(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"app", "--help"}
	if err := Run(&runWithHelpCmd{}); err != nil {
		t.Fatalf("Run with --help failed: %v", err)
	}
}

func TestRun_ShortHelp(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"app", "-h"}
	if err := Run(&runWithHelpCmd{}); err != nil {
		t.Fatalf("Run with -h failed: %v", err)
	}
}

func TestRun_BeforeError(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	cmd := &beforeErrorCmd{ShouldFail: true}
	os.Args = []string{"app"}
	if err := Run(cmd); err == nil {
		t.Error("expected error from Before")
	}
}

func TestRun_UnknownCommand(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"app", "unknown"}
	if err := Run(&testRootCmd{}); err == nil {
		t.Error("Run with unknown command should return a non-nil error")
	}
}

type rootWithBeforeAndAfter struct {
	BeforeCalled bool
	AfterCalled  bool
	RunCalled    bool
}

func (c *rootWithBeforeAndAfter) Before() error {
	c.BeforeCalled = true
	return nil
}

func (c *rootWithBeforeAndAfter) Run() error {
	c.RunCalled = true
	return nil
}

func (c *rootWithBeforeAndAfter) After() error {
	c.AfterCalled = true
	return nil
}

func TestRun_RootLifecycle(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	cmd := &rootWithBeforeAndAfter{}
	os.Args = []string{"app"}
	if err := Run(cmd); err != nil {
		t.Fatalf("Run failed: %v", err)
	}
	if !cmd.BeforeCalled {
		t.Error("expected Before to be called")
	}
	if !cmd.RunCalled {
		t.Error("expected Run to be called")
	}
	if !cmd.AfterCalled {
		t.Error("expected After to be called")
	}
}

type afterErrorCmd struct{}

func (c *afterErrorCmd) Before() error { return nil }
func (c *afterErrorCmd) Run() error    { return nil }
func (c *afterErrorCmd) After() error  { return errAfterFailed }

func TestRun_AfterError(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"app"}
	if err := Run(&afterErrorCmd{}); err == nil {
		t.Error("expected error from After")
	}
}

func TestNew_WithOptions(t *testing.T) {
	app, err := New(&testRootCmd{},
		WithVersion("1.0.0"),
		WithPanicRecovery(),
	)
	if err != nil {
		t.Fatalf("New with options failed: %v", err)
	}
	if app.version != "1.0.0" {
		t.Errorf("got version %q, want %q", app.version, "1.0.0")
	}
	if !app.panicRecovery {
		t.Error("expected panicRecovery to be true")
	}
}

func TestNew_WithContainer(t *testing.T) {
	builder := di.NewBuilder()
	container, err := builder.Build()
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	app, err := New(&testRootCmd{}, WithContainer(container))
	if err != nil {
		t.Fatalf("New with container failed: %v", err)
	}
	if app.Container == nil {
		t.Error("expected non-nil Container")
	}
}

func TestNew_WithValidator(t *testing.T) {
	v := validation.New()
	app, err := New(&testRootCmd{}, WithValidator(foundationValidator(v)))
	if err != nil {
		t.Fatalf("New with validator failed: %v", err)
	}
	if app.Validator == nil {
		t.Error("expected non-nil Validator")
	}
}

type validationCmd struct {
	Email string `cli:"email" help:"Email address" validate:"email"`
}

func (c *validationCmd) Run() error { return nil }

func TestRun_WithValidation(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	v := validation.New()
	os.Args = []string{"app", "--email", "not-an-email"}
	err := Run(&validationCmd{}, WithValidator(foundationValidator(v)))
	if err == nil {
		t.Error("expected validation error")
	}
}

func TestRun_WithValidationPass(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	v := validation.New()
	os.Args = []string{"app", "--email", "user@example.com"}
	err := Run(&validationCmd{}, WithValidator(foundationValidator(v)))
	if err != nil {
		t.Errorf("expected no validation error, got: %v", err)
	}
}

type panicCmd struct{}

func (c *panicCmd) Run() error {
	panic("something went wrong")
}

func TestRun_WithPanicRecovery(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"app"}
	err := Run(&panicCmd{}, WithPanicRecovery())
	if err == nil {
		t.Error("expected error from panic recovery")
	}
}

func TestRun_WithoutPanicRecovery(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic to propagate without recovery")
		}
	}()
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"app"}
	Run(&panicCmd{})
}

func TestResolveCommand_PositionalAllowed(t *testing.T) {
	root := parser.NewCommandNode("app", "", reflect.ValueOf(struct{}{}))
	leaf := parser.NewCommandNode("greet", "", reflect.ValueOf(struct{}{}))
	leaf.Args = []*parser.ArgMetadata{{Description: "name"}}
	root.Children["greet"] = leaf

	n, rem, err := resolveCommand(root, []string{"greet", "alice"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.Name != "greet" {
		t.Fatalf("expected greet node, got %q", n.Name)
	}
	if len(rem) != 1 || rem[0] != "alice" {
		t.Fatalf("expected positional [alice], got %v", rem)
	}
}

type testTypedCmd struct {
	Threshold float64  `cli:"threshold" default:"0.6" help:"A float"`
	Ratio     float32  `cli:"ratio" help:"A 32 bit float"`
	Count     uint     `cli:"count" help:"An unsigned int"`
	If        []string `cli:"if" help:"A repeatable premise"`
	Name      string   `cli:"name" help:"A scalar"`
}

func (c *testTypedCmd) Run() error { return nil }

// bindFlags runs argv through the real parse and bind path.
func bindFlags(t *testing.T, cmd any, argv []string) {
	t.Helper()
	node, err := parser.Parse("root", cmd)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	parsed, pos, err := parseArgs(argv, node.Flags)
	if err != nil {
		t.Fatalf("parseArgs: %v", err)
	}
	if err := applyBindings(node, parsed, pos, node.Flags); err != nil {
		t.Fatalf("applyBindings: %v", err)
	}
}

// A float or unsigned field used to appear in the help and then fail to bind,
// because applyBindings registered a handler only for the kinds it named.
func TestApplyBindings_FloatAndUintFlags(t *testing.T) {
	cmd := &testTypedCmd{}
	bindFlags(t, cmd, []string{"--threshold", "0.9", "--ratio", "0.25", "--count", "7"})

	if cmd.Threshold != 0.9 {
		t.Errorf("threshold: got %v, want 0.9", cmd.Threshold)
	}
	if cmd.Ratio != 0.25 {
		t.Errorf("ratio: got %v, want 0.25", cmd.Ratio)
	}
	if cmd.Count != 7 {
		t.Errorf("count: got %v, want 7", cmd.Count)
	}
}

func TestApplyBindings_FloatFlagUsesDefault(t *testing.T) {
	cmd := &testTypedCmd{}
	bindFlags(t, cmd, nil)

	if cmd.Threshold != 0.6 {
		t.Errorf("threshold: got %v, want the declared default 0.6", cmd.Threshold)
	}
}

// Repeating a flag used to keep only its last occurrence, which left a slice
// field unable to ever hold more than one value.
func TestApplyBindings_RepeatedFlagFillsSlice(t *testing.T) {
	cmd := &testTypedCmd{}
	bindFlags(t, cmd, []string{"--if", "a,b,c", "--if", "d,e,f"})

	want := []string{"a,b,c", "d,e,f"}
	if !reflect.DeepEqual(cmd.If, want) {
		t.Errorf("if: got %v, want %v", cmd.If, want)
	}
}

// A scalar keeps the conventional meaning of a repeated flag: the last wins.
func TestApplyBindings_RepeatedScalarFlagTakesLast(t *testing.T) {
	cmd := &testTypedCmd{}
	bindFlags(t, cmd, []string{"--name", "first", "--name", "second"})

	if cmd.Name != "second" {
		t.Errorf("name: got %q, want %q", cmd.Name, "second")
	}
}

func TestParseArgs_RepeatedFlagKeepsEveryOccurrence(t *testing.T) {
	flags := map[string]*parser.FlagMetadata{
		"if": {
			Name:  "if",
			Field: reflect.ValueOf([]string{}),
		},
	}
	parsed, _, err := parseArgs([]string{"--if", "one", "--if=two", "--if", "three"}, flags)
	if err != nil {
		t.Fatalf("parseArgs failed: %v", err)
	}

	want := []string{"one", "two", "three"}
	if !reflect.DeepEqual(parsed["if"], want) {
		t.Errorf("got %v, want %v", parsed["if"], want)
	}
}
