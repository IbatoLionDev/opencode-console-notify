//go:build darwin

// macOS backend: osascript display notification (built-in, no install).
// The AUMID identity is a Windows concept; EnsureAUMID is a no-op success
// on macOS (see aumid/registry_darwin.go).
package notifier

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/IbatoLionDev/opencode-console-notify/internal/aumid"
)

// osascriptApp is the app name shown by the notification center.
const osascriptApp = "OpenCode"

// buildOsascriptArgs renders the osascript argument vector for one
// notification with title and body.
func buildOsascriptArgs(title, body string) []string {
	script := `display notification ` + strLiteral(body) + ` with title ` + strLiteral(title)
	return []string{"-e", script}
}

// strLiteral wraps a string as a quoted AppleScript string literal,
// escaping backslashes and double quotes so the payload can never break.
func strLiteral(s string) string {
	esc := strings.ReplaceAll(s, "\\", "\\\\")
	esc = strings.ReplaceAll(esc, "\"", "\\\"")
	return `"` + esc + `"`
}

// SendTestToast ensures the identity (no-op on macOS) and then shows a
// real notification so a manual run visibly notifies. osascript is
// built-in on macOS; a missing binary is a clear error.
func SendTestToast(reg aumid.Registry) error {
	if err := reg.EnsureAUMID(); err != nil {
		return err
	}
	if _, err := exec.LookPath("osascript"); err != nil {
		return fmt.Errorf("osascript not found (macOS built-in): %w", err)
	}
	args := buildOsascriptArgs("OpenCode", "Test notification from opencode-notify")
	cmd := exec.Command("osascript", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("show test toast: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
