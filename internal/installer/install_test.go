package installer

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/IbatoLionDev/opencode-console-notify/internal/aumid"
	"github.com/IbatoLionDev/opencode-console-notify/internal/config"
	"github.com/IbatoLionDev/opencode-console-notify/plugin"
)

var errTestBoom = errors.New("test registry failure")

func TestInstallWritesEmbeddedPluginAtomically(t *testing.T) {
	dir := t.TempDir()
	plugins := filepath.Join(dir, "plugins")
	reg := &aumid.FakeRegistry{}

	var out bytes.Buffer
	if err := Install(plugins, reg, &out); err != nil {
		t.Fatalf("Install failed: %v", err)
	}

	got, err := os.ReadFile(config.PluginPath(plugins))
	if err != nil {
		t.Fatalf("read installed file: %v", err)
	}
	if !bytes.Equal(got, plugin.Source) {
		t.Fatal("installed file differs from embedded plugin source")
	}

	// No temp-file litter may remain after the atomic write.
	leftovers, err := filepath.Glob(filepath.Join(plugins, "*.tmp"))
	if err != nil {
		t.Fatalf("glob temp files: %v", err)
	}
	if len(leftovers) != 0 {
		t.Fatalf("leftover temp files: %v", leftovers)
	}

	if !reg.Registered {
		t.Fatal("Install did not register the AUMID")
	}
	for _, want := range []string{"Installed:", "SHA256:", "AUMID:"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("install output missing %q:\n%s", want, out.String())
		}
	}
}

func TestInstallIsIdempotent(t *testing.T) {
	plugins := t.TempDir()
	reg := &aumid.FakeRegistry{}

	var first, second bytes.Buffer
	if err := Install(plugins, reg, &first); err != nil {
		t.Fatalf("first Install failed: %v", err)
	}
	before, err := os.ReadFile(config.PluginPath(plugins))
	if err != nil {
		t.Fatalf("read after first install: %v", err)
	}
	if err := Install(plugins, reg, &second); err != nil {
		t.Fatalf("second Install failed: %v", err)
	}
	after, err := os.ReadFile(config.PluginPath(plugins))
	if err != nil {
		t.Fatalf("read after second install: %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Fatal("re-install changed the plugin file content")
	}
	if reg.EnsureCalls != 2 {
		t.Fatalf("expected 2 EnsureAUMID calls, got %d", reg.EnsureCalls)
	}
}

func TestInstallPropagatesRegistryError(t *testing.T) {
	plugins := t.TempDir()
	reg := &aumid.FakeRegistry{EnsureErr: errTestBoom}
	if err := Install(plugins, reg, &bytes.Buffer{}); err == nil {
		t.Fatal("expected registry error, got nil")
	}
}

func TestUninstallRemovesOnlyItsOwnFile(t *testing.T) {
	plugins := t.TempDir()
	keep := filepath.Join(plugins, "other-plugin.js")
	if err := os.WriteFile(keep, []byte("// bystander"), 0o644); err != nil {
		t.Fatalf("seed bystander file: %v", err)
	}
	if err := os.WriteFile(config.PluginPath(plugins), plugin.Source, 0o644); err != nil {
		t.Fatalf("seed plugin file: %v", err)
	}
	reg := &aumid.FakeRegistry{Registered: true, DisplayName: aumid.AUMIDDisplayName}

	var out bytes.Buffer
	if err := Uninstall(plugins, reg, &out); err != nil {
		t.Fatalf("Uninstall failed: %v", err)
	}

	if _, err := os.Stat(config.PluginPath(plugins)); !os.IsNotExist(err) {
		t.Fatal("plugin file was not removed")
	}
	if _, err := os.Stat(keep); err != nil {
		t.Fatalf("bystander file must survive uninstall: %v", err)
	}
	if reg.Registered {
		t.Fatal("AUMID key was not removed")
	}
	if reg.RemovedParent {
		t.Fatal("uninstall must never touch parent keys")
	}
	if !strings.Contains(out.String(), "Uninstall complete") {
		t.Fatalf("uninstall output missing completion line:\n%s", out.String())
	}
}

func TestUninstallIsIdempotentWhenNothingInstalled(t *testing.T) {
	plugins := t.TempDir()
	reg := &aumid.FakeRegistry{}

	var out bytes.Buffer
	if err := Uninstall(plugins, reg, &out); err != nil {
		t.Fatalf("Uninstall of empty state failed: %v", err)
	}
	if !strings.Contains(out.String(), "nothing to remove") {
		t.Fatalf("expected nothing-to-remove messaging:\n%s", out.String())
	}
}

func TestUninstallPropagatesRegistryError(t *testing.T) {
	plugins := t.TempDir()
	reg := &aumid.FakeRegistry{RemoveErr: errTestBoom}
	if err := Uninstall(plugins, reg, &bytes.Buffer{}); err == nil {
		t.Fatal("expected registry error, got nil")
	}
}
