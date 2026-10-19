// Package notifier testexport exposes identifiers needed by black-box tests
// living in test/internal/notifier. Production code must not depend on this
// file; it exists only so the test package can probe internals without
// being in package notifier.

package notifier

// EscapeToastXML exposes escapeToastXML for black-box tests.
func EscapeToastXML(value string) string { return escapeToastXML(value) }
