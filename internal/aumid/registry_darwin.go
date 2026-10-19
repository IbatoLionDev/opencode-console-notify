//go:build darwin

package aumid

// registry_darwin.go is the no-op macOS Registry backend. macOS has no
// notification identity registry: osascript addresses the app by name, so
// every contract method succeeds without storing anything (mirror of
// registry_linux.go).

type darwinRegistry struct{}

// NewRegistry returns the no-op macOS backend.
func NewRegistry() Registry { return darwinRegistry{} }

// EnsureAUMID is a no-op success: there is nothing to register.
func (darwinRegistry) EnsureAUMID() error { return nil }

// AUMIDStatus always reports the healthy end state: the identity is the
// app name itself.
func (darwinRegistry) AUMIDStatus() (bool, string, error) {
	return true, "identity is the app name " + AUMID + " (no registry on macOS)", nil
}

// RemoveAUMID reports nothing to remove: there was never a key.
func (darwinRegistry) RemoveAUMID() (bool, error) { return false, nil }
