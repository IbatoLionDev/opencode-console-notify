//go:build linux

// Linux backend: notify-send from libnotify (argv is passed exact, no
// shell). The AUMID identity is a Windows concept; EnsureAUMID is a
// no-op success on Linux (see aumid/registry_linux.go).
package notifier

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/IbatoLionDev/opencode-console-notify/internal/aumid"
)

// notifySendApp is the --app-name shown by the notification daemon.
const notifySendApp = "OpenCode"

// buildNotifySendArgs renders the notify-send argument vector for one
// summary plus one body line.
func buildNotifySendArgs(title, line string) []string {
	return []string{"--app-name=" + notifySendApp, "--urgency=normal", "--expire-time=8000", title, line}
}

// SendTestToast ensures the identity (no-op on Linux) and then shows a
// real toast so a manual run visibly notifies. A missing notify-send
// is a clear error pointing at libnotify-bin.
func SendTestToast(reg aumid.Registry) error {
	if err := reg.EnsureAUMID(); err != nil {
		return err
	}
	if _, err := exec.LookPath("notify-send"); err != nil {
		return fmt.Errorf("notify-send not found (install libnotify-bin): %w", err)
	}
	args := buildNotifySendArgs("OpenCode", "Test notification from opencode-notify")
	cmd := exec.Command("notify-send", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("show test toast: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
