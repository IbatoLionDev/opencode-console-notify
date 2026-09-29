//go:build windows

package main

// registry_windows.go is the real HKCU-backed Registry implementation.
// HKCU-only: no admin rights are ever required.

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

// windowsRegistry implements Registry against HKCU.
type windowsRegistry struct{}

// NewRegistry returns the production HKCU registry backend.
func NewRegistry() Registry { return windowsRegistry{} }

// EnsureAUMID creates the AUMID key and DisplayName value when missing.
// The existence guards make a re-run a no-op, exactly like install.ps1
// and the plugin's ensureAumidRegistered().
func (windowsRegistry) EnsureAUMID() error {
	key, _, err := registry.CreateKey(
		registry.CURRENT_USER, AUMIDKeyPath, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return fmt.Errorf("create AUMID key: %w", err)
	}
	defer key.Close()

	val, _, err := key.GetStringValue("DisplayName")
	if err != nil || val == "" {
		if err := key.SetStringValue("DisplayName", AUMIDDisplayName); err != nil {
			return fmt.Errorf("set AUMID DisplayName: %w", err)
		}
	}
	return nil
}

// AUMIDStatus reports whether the AUMID end state is present: the exact
// key exists and carries a DisplayName value.
func (windowsRegistry) AUMIDStatus() (bool, string, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, AUMIDKeyPath, registry.QUERY_VALUE)
	if err != nil {
		return false, "AUMID registry key is missing", nil
	}
	defer key.Close()

	val, _, err := key.GetStringValue("DisplayName")
	if err != nil || val == "" {
		return false, "AUMID registry key exists but DisplayName is missing", nil
	}
	return true, fmt.Sprintf("registered (HKCU:\\Software\\Classes\\AppUserModelId\\%s, DisplayName=%s)", AUMID, val), nil
}

// RemoveAUMID deletes exactly the AUMID key, never its parent keys.
// A missing key is not an error.
func (windowsRegistry) RemoveAUMID() (bool, error) {
	err := registry.DeleteKey(registry.CURRENT_USER, AUMIDKeyPath)
	if err != nil {
		if err == registry.ErrNotExist {
			return false, nil
		}
		return false, fmt.Errorf("remove AUMID key: %w", err)
	}
	return true, nil
}
