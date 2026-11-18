package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/mirkobrombin/go-cli-builder/v3/pkg/parser"
)

func TestFishCompletion(t *testing.T) {
	fish, err := exec.LookPath("fish")
	if err != nil {
		t.Skip("fish is not installed")
	}
	root := makeNode("completion-test", "", &struct{}{})
	root.Flags["profile"] = &parser.FlagMetadata{Description: "User's profile", Field: reflect.ValueOf("")}
	remote := makeNode("remote", "Manage user's remotes", &struct{}{})
	remote.Flags["verbose"] = &parser.FlagMetadata{Short: "v", Field: reflect.ValueOf(false)}
	remote.Children["add"] = makeNode("add", `Add a user's remote in C:\apps`, &struct{}{})
	root.Children["remote"] = remote
	root.Children["r"] = remote
	remote.Aliases = []string{"r"}
	root.Children["run"] = makeNode("run", "Run an app", &struct{}{})
	var script bytes.Buffer
	if err := (&App{RootNode: root}).GenFishCompletion(&script); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "completion.fish")
	if err := os.WriteFile(file, script.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	if output, err := exec.Command(fish, "--no-config", "--no-execute", file).CombinedOutput(); err != nil {
		t.Fatalf("invalid fish script: %v\n%s\n%s", err, output, script.String())
	}
	for _, tc := range []struct {
		line   string
		want   string
		absent string
	}{
		{"completion-test ", "remote", "add"},
		{"completion-test --profile remote ", "remote", "add"},
		{"completion-test remote ", "add", "run"},
		{"completion-test r ", "add", "run"},
		{"completion-test remote -v ", "add", "run"},
		{"completion-test --pro", "--profile", "add"},
		{"completion-test remote --v", "--verbose", "--profile"},
		{"completion-test run remote ", "", "add"},
		{"completion-test remote add ", "", "add"},
	} {
		t.Run(tc.line, func(t *testing.T) {
			cmd := exec.Command(fish, "--no-config", "-c", `source $argv[1]; complete -C $argv[2]`, file, tc.line)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			output, err := cmd.Output()
			if err != nil {
				t.Fatalf("complete: %v\n%s", err, stderr.String())
			}
			if stderr.Len() != 0 {
				t.Fatalf("completion wrote to stderr: %s", stderr.String())
			}
			found := map[string]bool{}
			for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
				found[strings.SplitN(line, "\t", 2)[0]] = true
			}
			if tc.want != "" && !found[tc.want] {
				t.Errorf("missing %q in %s", tc.want, output)
			}
			if found[tc.absent] {
				t.Errorf("unexpected %q in %s", tc.absent, output)
			}
		})
	}
}
