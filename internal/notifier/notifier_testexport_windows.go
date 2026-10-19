//go:build windows

// Package notifier testexport exposes identifiers needed by black-box tests
// living in test/internal/notifier. Production code must not depend on this
// file; it exists only so the test package can probe internals without
// being in package notifier.

package notifier

// BuildToastScript exposes buildToastScript for black-box tests.
func BuildToastScript(title, line string) string { return buildToastScript(title, line) }

// EncodePowerShell exposes encodePowerShell for black-box tests.
func EncodePowerShell(script string) string { return encodePowerShell(script) }
