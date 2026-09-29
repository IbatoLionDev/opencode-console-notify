package main

// paths.go resolves where the plugin file lives.

import (
	"os"
	"path/filepath"
	"strings"
)

// PluginFileName is the only file this CLI ever writes or removes.
const PluginFileName = "console-notify.js"

// DefaultPluginsDir returns OpenCode's Windows plugin directory:
// <HOME>/.config/opencode/plugins. It mirrors install.ps1's default.
func DefaultPluginsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return "", errHomeNotFound
	}
	return filepath.Join(home, ".config", "opencode", "plugins"), nil
}

// ResolvePluginsDir returns the --plugins-dir override when non-blank,
// otherwise the default plugins directory.
func ResolvePluginsDir(override string) (string, error) {
	if strings.TrimSpace(override) != "" {
		return filepath.Clean(override), nil
	}
	return DefaultPluginsDir()
}

// PluginPath joins the plugins directory with the plugin file name.
func PluginPath(pluginsDir string) string {
	return filepath.Join(pluginsDir, PluginFileName)
}
