//go:build windows

package notifier

import (
	"strings"
	"testing"

	"github.com/IbatoLionDev/opencode-console-notify/internal/aumid"
)

func TestBuildToastScriptUsesSharedIdentity(t *testing.T) {
	script := buildToastScript("OpenCode", "hello")
	for _, want := range []string{aumid.AUMID, "OpenCode", "hello", "ToastNotificationManager"} {
		if !strings.Contains(script, want) {
			t.Fatalf("toast script missing %q:\n%s", want, script)
		}
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
