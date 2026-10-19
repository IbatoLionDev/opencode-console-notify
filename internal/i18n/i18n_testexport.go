// Package i18n testexport exposes identifiers needed by black-box tests
// living in test/internal/i18n. Production code must not depend on this
// file; it exists only so the test package can probe internals without
// being in package i18n.

package i18n

// En exposes the English dictionary for black-box tests.
var En = en

// Es exposes the Spanish dictionary for black-box tests.
var Es = es
