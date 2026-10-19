// Package installer testexport exposes identifiers needed by black-box
// tests living in test/internal/installer. Production code must not
// depend on this file; it exists only so the test package can probe
// internals without being in package installer.

package installer

// IdentityInstalledLine exposes identityInstalledLine for black-box tests.
func IdentityInstalledLine(goos string) string { return identityInstalledLine(goos) }

// IdentityRemovedLine exposes identityRemovedLine for black-box tests.
func IdentityRemovedLine(goos string, removed bool) string { return identityRemovedLine(goos, removed) }
