package fdn

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/mirkobrombin/go-cli-builder/v3/pkg/log"
	flogger "github.com/mirkobrombin/go-foundation/v2/core/logger"
)

func TestNewStructured_EmitsOneJSONObjectPerLine(t *testing.T) {
	var buf bytes.Buffer
	l := NewStructured(&buf)

	l.Info("started")
	l.Error("failed")

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), buf.String())
	}

	for _, line := range lines {
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatalf("line is not JSON: %q (%v)", line, err)
		}
		if entry["msg"] == nil {
			t.Errorf("entry has no message: %q", line)
		}
	}
}

func TestNewStructured_KeepsTheSuccessSeverity(t *testing.T) {
	var buf bytes.Buffer
	NewStructured(&buf).Success("done")

	var entry map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &entry); err != nil {
		t.Fatalf("not JSON: %q (%v)", buf.String(), err)
	}

	fields, ok := entry["fields"].(map[string]any)
	if !ok {
		t.Fatalf("entry carries no fields: %q", buf.String())
	}
	if fields["status"] != "success" {
		t.Errorf("success severity lost, fields were %v", fields)
	}
}

func TestWrap_RoundTrips(t *testing.T) {
	var buf bytes.Buffer
	inner := flogger.New(
		flogger.WithoutDefaultSink(),
		flogger.WithSink(flogger.NewConsoleSink(&buf)),
	)

	wrapped := Wrap(inner)
	wrapped.Info("hello")

	if !strings.Contains(buf.String(), "hello") {
		t.Errorf("message did not reach the wrapped logger, got %q", buf.String())
	}

	back, ok := Unwrap(wrapped)
	if !ok {
		t.Fatal("Unwrap failed on a logger this package produced")
	}
	if back != inner {
		t.Error("Unwrap returned a different logger")
	}
}

func TestWrap_NilFallsBackToConsole(t *testing.T) {
	if Wrap(nil) == nil {
		t.Error("Wrap(nil) must return a usable logger")
	}
}

// The bridge must keep satisfying the CLI contract.
var _ log.Logger = (*wrapped)(nil)
