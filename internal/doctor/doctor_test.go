package doctor

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/IbatoLionDev/opencode-console-notify/internal/aumid"
	"github.com/IbatoLionDev/opencode-console-notify/internal/config"
	"github.com/IbatoLionDev/opencode-console-notify/internal/installer"
	"github.com/IbatoLionDev/opencode-console-notify/internal/version"
	"github.com/IbatoLionDev/opencode-console-notify/plugin"
)

var errTestBoom = errors.New("test registry failure")

func seedPluginFileForHealth(t *testing.T, plugins string, seedFile, seedEmpty bool) {
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

func assertHealthReport(t *testing.T, report HealthReport, wantHealthy, wantPlugin, wantAUMID bool) {
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

func TestCheckHealthParityMatrix(t *testing.T) {
	tests := []struct {
		name        string
		seedFile    bool
		seedEmpty   bool
		registered  bool
		wantHealthy bool
		wantPlugin  bool
		wantAUMID   bool
	}{
		{name: "nothing installed", wantHealthy: false},
		{name: "file only", seedFile: true, wantHealthy: false, wantPlugin: true},
		{name: "AUMID only", registered: true, wantHealthy: false, wantAUMID: true},
		{name: "both present is healthy", seedFile: true, registered: true, wantHealthy: true, wantPlugin: true, wantAUMID: true},
		{name: "empty file counts as missing", seedFile: true, seedEmpty: true, registered: true, wantHealthy: false, wantAUMID: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugins := t.TempDir()
			seedPluginFileForHealth(t, plugins, tt.seedFile, tt.seedEmpty)
			reg := &aumid.FakeRegistry{Registered: tt.registered, DisplayName: aumid.AUMIDDisplayName}

			report, err := CheckHealth(plugins, reg)
			if err != nil {
				t.Fatalf("CheckHealth failed: %v", err)
			}
			assertHealthReport(t, report, tt.wantHealthy, tt.wantPlugin, tt.wantAUMID)
		})
	}
}

func TestDoctorExitCodesAndMessages(t *testing.T) {
	plugins := t.TempDir()
	reg := &aumid.FakeRegistry{}

	var missing bytes.Buffer
	if code := Doctor(plugins, reg, &missing); code == 0 {
		t.Fatal("doctor on empty state must exit non-zero")
	} else if code != 1 {
		t.Fatalf("doctor on empty state exited %d, want 1", code)
	}
	if !strings.Contains(missing.String(), "MISSING") {
		t.Fatalf("doctor must say what is missing:\n%s", missing.String())
	}

	if err := installer.Install(plugins, reg, &bytes.Buffer{}); err != nil {
		t.Fatalf("Install failed: %v", err)
	}
	var healthy bytes.Buffer
	if code := Doctor(plugins, reg, &healthy); code != 0 {
		t.Fatalf("doctor after install exited %d, want 0:\n%s", code, healthy.String())
	}
	if !strings.Contains(healthy.String(), "Healthy") {
		t.Fatalf("healthy doctor must say so:\n%s", healthy.String())
	}
}

func TestDoctorReportsRegistryFailure(t *testing.T) {
	plugins := t.TempDir()
	reg := &aumid.FakeRegistry{StatusErr: errTestBoom}
	if code := Doctor(plugins, reg, &bytes.Buffer{}); code == 0 {
		t.Fatal("doctor with registry failure must exit non-zero")
	}
}

func TestDoctorAgainstMissingDirectory(t *testing.T) {
	plugins := filepath.Join(t.TempDir(), "does-not-exist")
	reg := &aumid.FakeRegistry{}
	report, err := CheckHealth(plugins, reg)
	if err != nil {
		t.Fatalf("CheckHealth failed: %v", err)
	}
	if report.Healthy() || report.PluginOK {
		t.Fatalf("missing directory must not be healthy: %+v", report)
	}
}

func seedHealthyInstall(t *testing.T, plugins string, reg *aumid.FakeRegistry) {
	t.Helper()
	if err := installer.Install(plugins, reg, &bytes.Buffer{}); err != nil {
		t.Fatalf("Install failed: %v", err)
	}
}

func swapRegistryURL(t *testing.T, url string) {
	t.Helper()
	old := registryURL
	registryURL = url
	t.Cleanup(func() { registryURL = old })
}

func TestDoctorPrintsUpdateNoticeWhenNewer(t *testing.T) {
	plugins := t.TempDir()
	reg := &aumid.FakeRegistry{}
	seedHealthyInstall(t, plugins, reg)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"version":"9.9.9"}`))
	}))
	defer server.Close()
	swapRegistryURL(t, server.URL)

	var out bytes.Buffer
	if code := Doctor(plugins, reg, &out); code != 0 {
		t.Fatalf("healthy doctor exited %d, want 0", code)
	}
	want := "Update available: " + version.Version + " -> 9.9.9"
	if !strings.Contains(out.String(), want) {
		t.Fatalf("doctor must print update notice %q, got:\n%s", want, out.String())
	}
}

func TestDoctorStaysSilentWhenRegistryDown(t *testing.T) {
	plugins := t.TempDir()
	reg := &aumid.FakeRegistry{}
	seedHealthyInstall(t, plugins, reg)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Empty on purpose: the handler must produce no usable response so
		// the test server simulates a dead registry (it is closed right away).
	}))
	url := server.URL
	server.Close()
	swapRegistryURL(t, url)

	var out bytes.Buffer
	if code := Doctor(plugins, reg, &out); code != 0 {
		t.Fatalf("healthy doctor with dead registry exited %d, want 0", code)
	}
	if strings.Contains(out.String(), "Update available") {
		t.Fatalf("doctor must stay silent when registry is down, got:\n%s", out.String())
	}
}
