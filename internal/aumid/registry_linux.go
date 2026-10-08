//go:build linux

package aumid

// registry_linux.go is the no-op Linux Registry backend. Linux has no
// notification identity registry: notify-send addresses the app by
// name, so every contract method succeeds without storing anything.

type linuxRegistry struct{}

// NewRegistry returns the no-op Linux backend.
func NewRegistry() Registry { return linuxRegistry{} }

// EnsureAUMID is a no-op success: there is nothing to register.
func (linuxRegistry) EnsureAUMID() error { return nil }

// AUMIDStatus always reports the healthy end state: the identity is
// the app name itself.
func (linuxRegistry) AUMIDStatus() (bool, string, error) {
	return true, "identity is the app name " + AUMID + " (no registry on Linux)", nil
}

// RemoveAUMID reports nothing to remove: there was never a key.
func (linuxRegistry) RemoveAUMID() (bool, error) { return false, nil }
