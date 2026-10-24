// Package doctor testexport exposes identifiers needed by black-box tests
// living in test/internal/doctor. Production code must not depend on this
// file; it exists only so the test package can probe internals without
// being in package doctor.

package doctor

import (
	"bytes"
	"errors"
	"os"
	"testing"

	"github.com/IbatoLionDev/opencode-console-notify/internal/aumid"
	"github.com/IbatoLionDev/opencode-console-notify/internal/config"
	"github.com/IbatoLionDev/opencode-console-notify/internal/installer"
	"github.com/IbatoLionDev/opencode-console-notify/plugin"
)

var errTestBoom = errors.New("test registry failure")

// StubProgramBaseName pins the running-binary base name reported in the
// STALE note so tests never depend on the test binary's own name.
func StubProgramBaseName(t *testing.T, name string) {
	t.Helper()
	old := ProgramBaseName
	ProgramBaseName = func() string { return name }
	t.Cleanup(func() { ProgramBaseName = old })
}

// RegistryURL exposes registryURL (declared in notice.go) for black-box
// tests that need to swap the update-notice endpoint.
var RegistryURL = registryURL

// StubNotifier points the backend probe at a fixed result for one test.
// Tests must never depend on the machine PATH: a Linux box without
// notify-send (or Windows without PowerShell on PATH) would flip them.
func StubNotifier(t *testing.T, path string, err error) {
	t.Helper()
	old := NotifierLookPath
	NotifierLookPath = func(string) (string, error) { return path, err }
	t.Cleanup(func() { NotifierLookPath = old })
}

// AssertHealthReport asserts the health report matches expectations.
func AssertHealthReport(t *testing.T, report HealthReport, wantHealthy, wantPlugin, wantAUMID bool) {
	t.Helper()
	if report.Healthy() != wantHealthy {
		t.Fatalf("Healthy() = %v, want %v (%+v)", report.Healthy(), wantHealthy, report)
	}
	if report.PluginOK != wantPlugin {
		t.Fatalf("PluginOK = %v, want %v", report.PluginOK, wantPlugin)
	}
	if report.AUMIDOK != wantAUMID {
		t.Fatalf("AUMIDOK = %v, want %v", report.AUMIDOK, wantAUMID)
	}
}

// SeedPluginFileForHealth seeds the plugin file for health tests.
func SeedPluginFileForHealth(t *testing.T, plugins string, seedFile, seedEmpty bool) {
	t.Helper()
	if !seedFile {
		return
	}
	content := plugin.Source
	if seedEmpty {
		content = []byte{}
	}
	if err := os.WriteFile(config.PluginPath(plugins), content, 0o644); err != nil {
		t.Fatalf("seed plugin file: %v", err)
	}
}

// SeedHealthyInstall runs installer.Install and seeds a healthy install.
func SeedHealthyInstall(t *testing.T, plugins string, reg *aumid.FakeRegistry) {
	t.Helper()
	if err := installer.Install(plugins, reg, &bytes.Buffer{}); err != nil {
		t.Fatalf("Install failed: %v", err)
	}
}

// SwapRegistryURL swaps the registry URL for update-notice tests.
func SwapRegistryURL(t *testing.T, url string) {
	t.Helper()
	old := RegistryURL
	RegistryURL = url
	t.Cleanup(func() { RegistryURL = old })
}
