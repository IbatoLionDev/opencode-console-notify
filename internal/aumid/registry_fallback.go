//go:build !windows && !linux

package aumid

// registry_fallback.go is the stub for platforms with no identity
// backend yet (macOS is next). The CLI keeps compiling everywhere;
// only Windows and Linux install and notify today.

import "errors"

var errWindowsOnly = errors.New("registry access requires Windows")

type unsupportedRegistry struct{}

// NewRegistry returns a backend that always fails outside Windows.
func NewRegistry() Registry { return unsupportedRegistry{} }

func (unsupportedRegistry) EnsureAUMID() error { return errWindowsOnly }

func (unsupportedRegistry) AUMIDStatus() (bool, string, error) { return false, "", errWindowsOnly }

func (unsupportedRegistry) RemoveAUMID() (bool, error) { return false, errWindowsOnly }
