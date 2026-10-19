//go:build windows

// Package screen testexport exposes identifiers needed by black-box tests
// living in test/internal/screen. Production code must not depend on this
// file; it exists only so the test package can probe internals without
// being in package screen.

package screen

import "os"

// GetMode exposes getMode for the windows console test.
func GetMode(f *os.File) (uint32, error) { return getMode(f) }
