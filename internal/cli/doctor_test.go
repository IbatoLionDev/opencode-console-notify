package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
			if tt.seedFile {
				content := pluginSource
				if tt.seedEmpty {
					content = []byte{}
				}
				if err := os.WriteFile(PluginPath(plugins), content, 0o644); err != nil {
					t.Fatalf("seed plugin file: %v", err)
				}
			}
			reg := &FakeRegistry{registered: tt.registered, displayName: AUMIDDisplayName}

			report, err := CheckHealth(plugins, reg)
			if err != nil {
				t.Fatalf("CheckHealth failed: %v", err)
			}
			if report.Healthy() != tt.wantHealthy {
				t.Fatalf("Healthy() = %v, want %v (%+v)", report.Healthy(), tt.wantHealthy, report)
			}
			if report.PluginOK != tt.wantPlugin {
				t.Fatalf("PluginOK = %v, want %v", report.PluginOK, tt.wantPlugin)
			}
			if report.AUMIDOK != tt.wantAUMID {
				t.Fatalf("AUMIDOK = %v, want %v", report.AUMIDOK, tt.wantAUMID)
			}
		})
	}
}

func TestDoctorExitCodesAndMessages(t *testing.T) {
	plugins := t.TempDir()
	reg := &FakeRegistry{}

	var missing bytes.Buffer
	if code := Doctor(plugins, reg, &missing); code == 0 {
		t.Fatal("doctor on empty state must exit non-zero")
	} else if code != 1 {
		t.Fatalf("doctor on empty state exited %d, want 1", code)
	}
	if !strings.Contains(missing.String(), "MISSING") {
		t.Fatalf("doctor must say what is missing:\n%s", missing.String())
	}

	if err := Install(plugins, reg, &bytes.Buffer{}); err != nil {
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
	reg := &FakeRegistry{statusErr: errTestBoom}
	if code := Doctor(plugins, reg, &bytes.Buffer{}); code == 0 {
		t.Fatal("doctor with registry failure must exit non-zero")
	}
}

func TestDoctorAgainstMissingDirectory(t *testing.T) {
	plugins := filepath.Join(t.TempDir(), "does-not-exist")
	reg := &FakeRegistry{}
	report, err := CheckHealth(plugins, reg)
	if err != nil {
		t.Fatalf("CheckHealth failed: %v", err)
	}
	if report.Healthy() || report.PluginOK {
		t.Fatalf("missing directory must not be healthy: %+v", report)
	}
}

func TestRunDoctorEndToEnd(t *testing.T) {
	plugins := t.TempDir()
	reg := &FakeRegistry{}
	var out, errOut bytes.Buffer

	if code := Run([]string{"doctor", "--plugins-dir", plugins}, reg, &out, &errOut); code == 0 {
		t.Fatal("run doctor before install must exit non-zero")
	}
	if code := Run([]string{"install", "--plugins-dir", plugins}, reg, &out, &errOut); code != 0 {
		t.Fatalf("run install exited %d: %s", code, errOut.String())
	}
	out.Reset()
	if code := Run([]string{"doctor", "--plugins-dir", plugins}, reg, &out, &errOut); code != 0 {
		t.Fatalf("run doctor after install exited %d:\n%s", code, out.String())
	}
	out.Reset()
	if code := Run([]string{"uninstall", "--plugins-dir", plugins}, reg, &out, &errOut); code != 0 {
		t.Fatalf("run uninstall exited %d: %s", code, errOut.String())
	}
}
