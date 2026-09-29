package cli

// registry.go declares the AUMID registry boundary.
//
// The Registry interface keeps every unit test hermetic: production code
// runs against HKCU on Windows, tests inject FakeRegistry (see the
// _test files) so nothing touches the real registry.

import "errors"

// AUMID is the Windows notification identity shared by the plugin,
// install.ps1, and this CLI.
const AUMID = "OpenCode.Notifier"

// AUMIDDisplayName is the friendly name stored on the AUMID key.
const AUMIDDisplayName = "OpenCode"

// AUMIDKeyPath is the HKCU-relative path of the exact key this CLI owns.
// Uninstall removes this key only, never its parents.
const AUMIDKeyPath = `Software\Classes\AppUserModelId\OpenCode.Notifier`

var errHomeNotFound = errors.New("could not determine the home directory")

// Registry abstracts the HKCU AUMID key owned by this tool.
type Registry interface {
	// EnsureAUMID creates the AUMID key and DisplayName value when
	// missing. Re-running must be a no-op.
	EnsureAUMID() error
	// AUMIDStatus reports whether the end state install.ps1 leaves
	// behind is present, with a human-readable detail string.
	AUMIDStatus() (registered bool, detail string, err error)
	// RemoveAUMID deletes exactly the AUMID key. It reports whether
	// the key existed; removing a missing key is not an error.
	RemoveAUMID() (removed bool, err error)
}
