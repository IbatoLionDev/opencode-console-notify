// Package notifier sends a real toast for the test command.
// Backends live per OS (toast_windows.go, toast_linux.go); shared
// payload helpers stay here so every backend renders text the same way.
package notifier

import (
	"strings"
)

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
