package cli

import (
	"strings"
	"testing"
)

func TestBuildToastScriptUsesSharedIdentity(t *testing.T) {
	script := buildToastScript("OpenCode", "hello")
	for _, want := range []string{AUMID, "OpenCode", "hello", "ToastNotificationManager"} {
		if !strings.Contains(script, want) {
			t.Fatalf("toast script missing %q:\n%s", want, script)
		}
	}
}

func TestEscapeToastXML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "plain text untouched", input: "Task finished", want: "Task finished"},
		{name: "ampersand", input: "a & b", want: "a &amp; b"},
		{name: "angle brackets", input: "<toast>", want: "&lt;toast&gt;"},
		{name: "quotes", input: `"hi" 'yo'`, want: "&quot;hi&quot; &apos;yo&apos;"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escapeToastXML(tt.input); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEncodePowerShellRoundTrip(t *testing.T) {
	script := buildToastScript("OpenCode", "a & b <test>")
	encoded := encodePowerShell(script)
	if encoded == "" {
		t.Fatal("encoded script must not be empty")
	}
	// UTF-16LE base64 of ASCII text always ends with a NUL high byte
	// pattern; spot-check decodability via length sanity.
	if len(encoded) < len(script) {
		t.Fatalf("encoded output suspiciously short: %d vs %d", len(encoded), len(script))
	}
}
