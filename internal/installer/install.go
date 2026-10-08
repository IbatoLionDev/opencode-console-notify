// Package installer implements the install and uninstall commands.
package installer

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/IbatoLionDev/opencode-console-notify/internal/aumid"
	"github.com/IbatoLionDev/opencode-console-notify/internal/config"
	"github.com/IbatoLionDev/opencode-console-notify/plugin"
)

// writeFileAtomic writes data to dest via a temp file in the same
// directory followed by a rename, so a running OpenCode never observes
// a half-written plugin file. Any leftover temp file is cleaned up.
func writeFileAtomic(dest string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(dest)
	tmp, err := os.CreateTemp(dir, config.PluginFileName+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	// Best effort cleanup: after a successful rename the temp path no
	// longer exists and Remove is a harmless no-op.
	defer os.Remove(tmpName)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmpName, dest); err != nil {
		return fmt.Errorf("rename temp file into place: %w", err)
	}
	return nil
}

// Install copies the embedded plugin into pluginsDir (creating it when
// needed) and idempotently registers the AUMID key.
func Install(pluginsDir string, reg aumid.Registry, out io.Writer) error {
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		return fmt.Errorf("create plugins directory: %w", err)
	}

	dest := config.PluginPath(pluginsDir)
	if err := writeFileAtomic(dest, plugin.Source, 0o644); err != nil {
		return err
	}

	if err := reg.EnsureAUMID(); err != nil {
		return err
	}

	sum := sha256.Sum256(plugin.Source)
	fmt.Fprintf(out, "Installed: %s\n", dest)
	fmt.Fprintf(out, "SHA256:    %x\n", sum)
	fmt.Fprintf(out, "Source:    embedded plugin/%s (no download)\n", config.PluginFileName)
	fmt.Fprint(out, identityInstalledLine(runtime.GOOS))
	fmt.Fprintf(out, "Restart OpenCode to load the plugin.\n")
	return nil
}

// identityInstalledLine reports the identity half of the install output
// for one OS: the HKCU key on Windows, the app name everywhere else.
// Taking goos as a parameter keeps the wording unit-testable on any OS.
func identityInstalledLine(goos string) string {
	if goos == "windows" {
		return fmt.Sprintf("AUMID:     registered (HKCU:\\Software\\Classes\\AppUserModelId\\%s, DisplayName=%s)\n", aumid.AUMID, aumid.AUMIDDisplayName)
	}
	return fmt.Sprintf("Identity:  app name %s (no registry on %s)\n", aumid.AUMID, goos)
}

// identityRemovedLine reports the identity half of the uninstall output
// for one OS and whether a marker was removed.
func identityRemovedLine(goos string, removed bool) string {
	if goos == "windows" {
		if removed {
			return fmt.Sprintf("Removed AUMID registry key: HKCU:\\Software\\Classes\\AppUserModelId\\%s\n", aumid.AUMID)
		}
		return fmt.Sprintf("AUMID registry key not present, nothing to remove: HKCU:\\Software\\Classes\\AppUserModelId\\%s\n", aumid.AUMID)
	}
	if removed {
		return fmt.Sprintf("Removed identity marker for %s (no registry on %s)\n", aumid.AUMID, goos)
	}
	return fmt.Sprintf("No identity marker for %s (no registry on %s), nothing to remove\n", aumid.AUMID, goos)
}

// Uninstall removes only the plugin file and exactly the AUMID key,
// never parent keys or anything else in the plugins directory.
func Uninstall(pluginsDir string, reg aumid.Registry, out io.Writer) error {
	dest := config.PluginPath(pluginsDir)
	if _, err := os.Stat(dest); err == nil {
		if err := os.Remove(dest); err != nil {
			return fmt.Errorf("remove plugin file: %w", err)
		}
		fmt.Fprintf(out, "Removed plugin file: %s\n", dest)
	} else if os.IsNotExist(err) {
		fmt.Fprintf(out, "Plugin file not present, nothing to remove: %s\n", dest)
	} else {
		return fmt.Errorf("stat plugin file: %w", err)
	}

	removed, err := reg.RemoveAUMID()
	if err != nil {
		return err
	}
	fmt.Fprint(out, identityRemovedLine(runtime.GOOS, removed))

	fmt.Fprintf(out, "Uninstall complete. Restart OpenCode to unload the plugin.\n")
	return nil
}
