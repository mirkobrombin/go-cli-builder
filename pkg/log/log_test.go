package log

import (
	"bytes"
	"strings"
	"testing"
)

// The console logger is what a user of the CLI actually sees, so these tests
// pin the two properties that matter: messages stay readable, and the four
// severities stay distinguishable.

func TestConsoleLogger_WritesReadableLines(t *testing.T) {
	var out bytes.Buffer
	l := NewWriter(&out, &out)

	l.Info("hello %s", "world")

	got := out.String()
	if !strings.Contains(got, "hello world") {
		t.Errorf("message not rendered, got %q", got)
	}
	if strings.Contains(got, "{") || strings.Contains(got, "\"msg\"") {
		t.Errorf("console output must not be JSON, got %q", got)
	}
}

func TestConsoleLogger_SeparatesOutAndErr(t *testing.T) {
	var out, errOut bytes.Buffer
	l := NewWriter(&out, &errOut)

	l.Info("info")
	l.Success("success")
	l.Warning("warning")
	l.Error("error")

	if !strings.Contains(out.String(), "info") || !strings.Contains(out.String(), "success") {
		t.Errorf("info and success belong on stdout, got %q", out.String())
	}
	if strings.Contains(out.String(), "warning") || strings.Contains(out.String(), "error") {
		t.Errorf("warning and error must not reach stdout, got %q", out.String())
	}
	if !strings.Contains(errOut.String(), "warning") || !strings.Contains(errOut.String(), "error") {
		t.Errorf("warning and error belong on stderr, got %q", errOut.String())
	}
}

func TestConsoleLogger_LevelsAreDistinguishable(t *testing.T) {
	render := func(fn func(Logger)) string {
		var buf bytes.Buffer
		fn(NewWriter(&buf, &buf))
		return buf.String()
	}

	info := render(func(l Logger) { l.Info("same text") })
	success := render(func(l Logger) { l.Success("same text") })

	if info == success {
		t.Error("Success must not be indistinguishable from Info")
	}
}

func TestConsoleLogger_NoColorWhenNotATerminal(t *testing.T) {
	var buf bytes.Buffer
	NewWriter(&buf, &buf).Error("boom")

	if strings.Contains(buf.String(), "\033[") {
		t.Errorf("ANSI codes must not be written to a non terminal, got %q", buf.String())
	}
}

func TestConsoleLogger_MessageWithoutArgsIsNotFormatted(t *testing.T) {
	var buf bytes.Buffer
	NewWriter(&buf, &buf).Info("100% done")

	if !strings.Contains(buf.String(), "100% done") {
		t.Errorf("a lone message must be printed verbatim, got %q", buf.String())
	}
}

func TestLogLevel_String(t *testing.T) {
	cases := map[LogLevel]string{
		LogLevelInfo:    "info",
		LogLevelSuccess: "success",
		LogLevelWarning: "warning",
		LogLevelError:   "error",
	}
	for level, want := range cases {
		if got := level.String(); got != want {
			t.Errorf("level %d: got %q, want %q", level, got, want)
		}
	}
}
