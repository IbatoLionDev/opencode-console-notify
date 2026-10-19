//go:build linux

// Package notifier testexport exposes identifiers needed by black-box tests
// living in test/internal/notifier. Production code must not depend on this
// file; it exists only so the test package can probe internals without
// being in package notifier.

package notifier

// BuildNotifySendArgs exposes buildNotifySendArgs for black-box tests.
func BuildNotifySendArgs(title, body string) []string { return buildNotifySendArgs(title, body) }
