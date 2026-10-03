//go:build !windows

package screen

import "errors"

var errWindowsOnly = errors.New("fullscreen console requires Windows")

// Enable always fails outside Windows so callers fall back to line mode.
func Enable() (func(), error) {
	return func() {}, errWindowsOnly
}

// Size reports a fixed fallback outside Windows.
func Size() (int, int) {
	return 80, 24
}
