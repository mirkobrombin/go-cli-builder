package help

import (
	"reflect"
	"strings"
	"testing"

	"github.com/mirkobrombin/go-cli-builder/v3/pkg/parser"
)

func newTestNode(name, desc string) *parser.CommandNode {
	return parser.NewCommandNode(name, desc, reflect.ValueOf(struct{}{}))
}

func TestGenerateHelp_RootOnly(t *testing.T) {
	node := newTestNode("app", "")
	out := GenerateHelp(node, nil)
	if !strings.Contains(out, "Usage: app [flags]") {
		t.Errorf("expected usage line, got: %s", out)
	}
}

func TestGenerateHelp_WithDescription(t *testing.T) {
	node := newTestNode("app", "My awesome CLI app")
	out := GenerateHelp(node, nil)
	if !strings.Contains(out, "My awesome CLI app") {
		t.Errorf("expected description in help, got: %s", out)
	}
}

func TestGenerateHelp_WithChildren(t *testing.T) {
	node := newTestNode("app", "")
	child := newTestNode("add", "Add a new item")
	node.Children["add"] = child

	out := GenerateHelp(node, nil)
	if !strings.Contains(out, "Commands:") {
		t.Errorf("expected Commands section, got: %s", out)
	}
	if !strings.Contains(out, "add") {
		t.Errorf("expected 'add' in help, got: %s", out)
	}
	if !strings.Contains(out, "Add a new item") {
		t.Errorf("expected child description, got: %s", out)
	}
}

func TestGenerateHelp_WithAliases(t *testing.T) {
	node := newTestNode("app", "")
	child := newTestNode("add", "Add")
	child.Aliases = []string{"a", "ad"}
	node.Children["add"] = child
	node.Children["a"] = child
	node.Children["ad"] = child

	out := GenerateHelp(node, nil)
	if !strings.Contains(out, "aliases:") {
		t.Errorf("expected aliases in help, got: %s", out)
	}
}

func TestGenerateHelp_WithFlags(t *testing.T) {
	node := newTestNode("app", "")
	node.Flags["verbose"] = &parser.FlagMetadata{
		Name:        "verbose",
		Short:       "v",
		Description: "Enable verbose",
		Field:       reflect.Value{},
	}

	out := GenerateHelp(node, nil)
	if !strings.Contains(out, "Flags:") {
		t.Errorf("expected Flags section, got: %s", out)
	}
	if !strings.Contains(out, "--verbose") {
		t.Errorf("expected --verbose, got: %s", out)
	}
	if !strings.Contains(out, "-v") {
		t.Errorf("expected -v shorthand, got: %s", out)
	}
	if !strings.Contains(out, "Enable verbose") {
		t.Errorf("expected flag description, got: %s", out)
	}
}

func TestGenerateHelp_WithFlagEnv(t *testing.T) {
	node := newTestNode("app", "")
	node.Flags["name"] = &parser.FlagMetadata{
		Name:  "name",
		Env:   "NAME_ENV",
		Field: reflect.Value{},
	}

	out := GenerateHelp(node, nil)
	if !strings.Contains(out, "env: NAME_ENV") {
		t.Errorf("expected env in details, got: %s", out)
	}
}

func TestGenerateHelp_WithFlagDefault(t *testing.T) {
	node := newTestNode("app", "")
	node.Flags["name"] = &parser.FlagMetadata{
		Name:    "name",
		Default: "world",
		Field:   reflect.Value{},
	}

	out := GenerateHelp(node, nil)
	if !strings.Contains(out, "default: world") {
		t.Errorf("expected default in details, got: %s", out)
	}
}

func TestGenerateHelp_WithFlagRequired(t *testing.T) {
	node := newTestNode("app", "")
	node.Flags["count"] = &parser.FlagMetadata{
		Name:     "count",
		Required: true,
		Field:    reflect.Value{},
	}

	out := GenerateHelp(node, nil)
	if !strings.Contains(out, "required") {
		t.Errorf("expected required in details, got: %s", out)
	}
}

func TestGenerateHelp_WithArgs(t *testing.T) {
	node := newTestNode("app", "")
	node.Args = append(node.Args, &parser.ArgMetadata{
		Description: "item",
		Required:    true,
	})

	out := GenerateHelp(node, nil)
	if !strings.Contains(out, "[item]") {
		t.Errorf("expected [item] in help, got: %s", out)
	}
}

func TestGenerateHelp_WithGreedyArgs(t *testing.T) {
	node := newTestNode("app", "")
	node.Args = append(node.Args, &parser.ArgMetadata{
		Description: "files",
		IsGreedy:    true,
	})

	out := GenerateHelp(node, nil)
	if !strings.Contains(out, "[files...]") {
		t.Errorf("expected [files...] in help, got: %s", out)
	}
}

func TestGenerateHelp_WithTranslator(t *testing.T) {
	node := newTestNode("app", "pr:app.desc")
	child := newTestNode("add", "pr:add.desc")
	node.Children["add"] = child

	tr := func(s string) string {
		m := map[string]string{
			"app.desc": "My App",
			"add.desc": "Add items",
		}
		return m[s]
	}

	out := GenerateHelp(node, tr)
	if !strings.Contains(out, "My App") {
		t.Errorf("expected translated desc, got: %s", out)
	}
	if !strings.Contains(out, "Add items") {
		t.Errorf("expected translated child desc, got: %s", out)
	}
}

func TestGenerateHelp_MultiFlag(t *testing.T) {
	node := newTestNode("app", "")
	node.Flags["verbose"] = &parser.FlagMetadata{
		Name:        "verbose",
		Short:       "v",
		Description: "Verbose output",
		Field:       reflect.Value{},
	}
	node.Flags["name"] = &parser.FlagMetadata{
		Name:        "name",
		Short:       "n",
		Description: "Your name",
		Default:     "world",
		Field:       reflect.Value{},
	}
	node.Flags["count"] = &parser.FlagMetadata{
		Name:        "count",
		Description: "Count things",
		Required:    true,
		Field:       reflect.Value{},
	}

	out := GenerateHelp(node, nil)
	if !strings.Contains(out, "--verbose") {
		t.Errorf("expected --verbose, got: %s", out)
	}
	if !strings.Contains(out, "--name") {
		t.Errorf("expected --name, got: %s", out)
	}
	if !strings.Contains(out, "--count") {
		t.Errorf("expected --count, got: %s", out)
	}
}
