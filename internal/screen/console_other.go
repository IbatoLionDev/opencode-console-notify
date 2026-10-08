//go:build !windows && !linux

package screen

import "errors"

var errUnsupportedConsole = errors.New("fullscreen console requires Windows or Linux")

// Enable always fails where no console backend exists so callers fall
// back to line mode.
func Enable() (func(), error) {
	return func() {
		// Empty on purpose: nothing was enabled, so nothing needs restoring.
	}, errUnsupportedConsole
}

// Size reports a fixed fallback where no console backend exists.
func Size() (int, int) {
	return 80, 24
}
