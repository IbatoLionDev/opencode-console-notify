//go:build !windows && !linux

// Stub for platforms with no toast backend yet (macOS is next): the
// package keeps compiling everywhere, sending stays a clear error.
package notifier

import (
	"errors"

	"github.com/IbatoLionDev/opencode-console-notify/internal/aumid"
)

var errUnsupportedPlatform = errors.New("test toast requires Windows or Linux")

// SendTestToast always fails outside Windows and Linux.
func SendTestToast(reg aumid.Registry) error { return errUnsupportedPlatform }
