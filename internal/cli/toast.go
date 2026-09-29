package cli

// toast.go sends a real Windows toast for the test command.
// It shells out to PowerShell with a UTF-16LE base64 payload, the same
// dependency-free technique the plugin uses, so quoting can never break.

import (
	"encoding/base64"
	"fmt"
	"os/exec"
	"strings"
	"unicode/utf16"
)

// buildToastScript renders the PowerShell snippet that shows one toast
// through the shared AUMID identity.
func buildToastScript(title, line string) string {
	xml := `<toast duration="long"><visual><binding template="ToastGeneric">` +
		`<text>` + escapeToastXML(title) + `</text>` +
		`<text>` + escapeToastXML(line) + `</text>` +
		`</binding></visual><audio src="ms-winsoundevent:Notification.Default"/></toast>`
	lines := []string{
		"[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null",
		"[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null",
		"$x = New-Object Windows.Data.Xml.Dom.XmlDocument",
		fmt.Sprintf(`$x.LoadXml('%s')`, xml),
		"$t = [Windows.UI.Notifications.ToastNotification]::new($x)",
		"$t.Tag = 'opencode-test'",
		"$t.Group = 'opencode'",
		fmt.Sprintf(`[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('%s').Show($t)`, AUMID),
	}
	return strings.Join(lines, "\n")
}

// escapeToastXML escapes the five XML metacharacters for toast payloads.
func escapeToastXML(value string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&apos;",
	)
	return replacer.Replace(value)
}

// encodePowerShell encodes a script as UTF-16LE base64 for
// powershell.exe -EncodedCommand, avoiding every shell-quoting pitfall.
func encodePowerShell(script string) string {
	units := utf16.Encode([]rune(script))
	raw := make([]byte, 0, len(units)*2)
	for _, u := range units {
		raw = append(raw, byte(u), byte(u>>8))
	}
	return base64.StdEncoding.EncodeToString(raw)
}

// SendTestToast ensures the AUMID registration exists (idempotent) and
// then shows a real toast so a manual run visibly notifies.
func SendTestToast(reg Registry) error {
	if err := reg.EnsureAUMID(); err != nil {
		return err
	}
	script := buildToastScript("OpenCode", "Test notification from opencode-notify")
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive",
		"-EncodedCommand", encodePowerShell(script))
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("show test toast: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}
