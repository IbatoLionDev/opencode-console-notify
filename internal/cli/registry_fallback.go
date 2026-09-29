//go:build !windows

package cli

// registry_fallback.go is the non-Windows Registry stub. The CLI is
// Windows-only in v1; this keeps the package compiling elsewhere.

import "errors"

var errWindowsOnly = errors.New("registry access requires Windows")

type unsupportedRegistry struct{}

// NewRegistry returns a backend that always fails outside Windows.
func NewRegistry() Registry { return unsupportedRegistry{} }

func (unsupportedRegistry) EnsureAUMID() error { return errWindowsOnly }

func (unsupportedRegistry) AUMIDStatus() (bool, string, error) { return false, "", errWindowsOnly }

func (unsupportedRegistry) RemoveAUMID() (bool, error) { return false, errWindowsOnly }
