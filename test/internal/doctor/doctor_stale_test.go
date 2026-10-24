// Black-box tests for the STALE plugin-file state: doctor mirrors the JS
// content check (lib/doctor/health.js) against the plugin copy embedded
// in this build, with MISSING states keeping wording and precedence.
package doctor_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/IbatoLionDev/opencode-console-notify/internal/aumid"
	"github.com/IbatoLionDev/opencode-console-notify/internal/config"
	"github.com/IbatoLionDev/opencode-console-notify/internal/doctor"
	"github.com/IbatoLionDev/opencode-console-notify/plugin"
)

// staleSeed returns modified copies of the embedded plugin bytes: same
// file shape, non-empty, but never equal to this build's Source. It must
// copy: plugin.Source is shared package state.
func staleSeed() []byte {
	seed := make([]byte, 0, len(plugin.Source)+14)
	seed = append(seed, plugin.Source...)
	return append(seed, "\n// stale bytes"...)
}

// seedPluginContent writes content at the resolved plugin file path.
func seedPluginContent(t *testing.T, plugins string, content []byte) {
	t.Helper()
	if err := os.WriteFile(config.PluginPath(plugins), content, 0o644); err != nil {
		t.Fatalf("seed plugin content: %v", err)
	}
}

func TestCheckHealthEqualBytesReportsOk(t *testing.T) {
	doctor.StubNotifier(t, "/bin/notify-send", nil)
	plugins := t.TempDir()
	reg := &aumid.FakeRegistry{Registered: true, DisplayName: aumid.AUMIDDisplayName}
	seedPluginContent(t, plugins, plugin.Source)

	report, err := doctor.CheckHealth(plugins, reg)
	if err != nil {
		t.Fatalf("CheckHealth failed: %v", err)
	}
	want := fmt.Sprintf("OK (%s)", config.PluginPath(plugins))
	if report.PluginNote != want {
		t.Fatalf("PluginNote = %q, want %q", report.PluginNote, want)
	}
	doctor.AssertHealthReport(t, report, true, true, true)
}

func TestCheckHealthStalePluginFileFails(t *testing.T) {
	doctor.StubNotifier(t, "/bin/notify-send", nil)
	doctor.StubProgramBaseName(t, "opencode-notify.exe")
	plugins := t.TempDir()
	reg := &aumid.FakeRegistry{Registered: true, DisplayName: aumid.AUMIDDisplayName}
	seedPluginContent(t, plugins, staleSeed())

	report, err := doctor.CheckHealth(plugins, reg)
	if err != nil {
		t.Fatalf("CheckHealth failed: %v", err)
	}
	want := fmt.Sprintf(
		"STALE (%s differs from this build; run \"opencode-notify.exe install\" to rewrite it)",
		config.PluginPath(plugins))
	if report.PluginNote != want {
		t.Fatalf("PluginNote = %q, want %q", report.PluginNote, want)
	}
	doctor.AssertHealthReport(t, report, false, false, true)
}

func TestStaleNoteResolvesProgramBaseName(t *testing.T) {
	doctor.StubNotifier(t, "/bin/notify-send", nil)
	// A name no branch hardcodes: only filepath.Base(os.Args[0]) with the
	// test stub resolves it.
	doctor.StubProgramBaseName(t, "opencode-notify-linux")
	plugins := t.TempDir()
	reg := &aumid.FakeRegistry{Registered: true, DisplayName: aumid.AUMIDDisplayName}
	seedPluginContent(t, plugins, staleSeed())

	report, err := doctor.CheckHealth(plugins, reg)
	if err != nil {
		t.Fatalf("CheckHealth failed: %v", err)
	}
	want := fmt.Sprintf(
		"STALE (%s differs from this build; run \"opencode-notify-linux install\" to rewrite it)",
		config.PluginPath(plugins))
	if report.PluginNote != want {
		t.Fatalf("PluginNote = %q, want %q", report.PluginNote, want)
	}
}

func TestCheckHealthMissingPrecedenceOverStale(t *testing.T) {
	doctor.StubNotifier(t, "/bin/notify-send", nil)
	doctor.StubProgramBaseName(t, "opencode-notify.exe")
	tests := []struct {
		name string
		want string
		seed func(t *testing.T, plugins string)
	}{
		{
			name: "directory keeps MISSING",
			want: "MISSING (%s is a directory, not a file)",
			seed: func(t *testing.T, plugins string) {
				if err := os.MkdirAll(plugins, 0o755); err != nil {
					t.Fatalf("create plugins dir: %v", err)
				}
				if err := os.Mkdir(config.PluginPath(plugins), 0o755); err != nil {
					t.Fatalf("create plugin dir: %v", err)
				}
			},
		},
		{
			name: "empty file keeps MISSING",
			want: "MISSING (%s is empty)",
			seed: func(t *testing.T, plugins string) {
				if err := os.MkdirAll(plugins, 0o755); err != nil {
					t.Fatalf("create plugins dir: %v", err)
				}
				seedPluginContent(t, plugins, []byte{})
			},
		},
		{
			name: "no file at all keeps MISSING",
			want: "MISSING (no file at %s)",
			seed: func(t *testing.T, plugins string) {
				if err := os.MkdirAll(plugins, 0o755); err != nil {
					t.Fatalf("create plugins dir: %v", err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugins := t.TempDir()
			tt.seed(t, plugins)
			reg := &aumid.FakeRegistry{Registered: true, DisplayName: aumid.AUMIDDisplayName}

			report, err := doctor.CheckHealth(plugins, reg)
			if err != nil {
				t.Fatalf("CheckHealth failed: %v", err)
			}
			want := fmt.Sprintf(tt.want, config.PluginPath(plugins))
			if report.PluginNote != want {
				t.Fatalf("PluginNote = %q, want %q", report.PluginNote, want)
			}
			if report.PluginOK {
				t.Fatalf("PluginOK = true, want false: %+v", report)
			}
		})
	}
}

func TestDoctorStalePluginExitsOne(t *testing.T) {
	doctor.StubNotifier(t, "/bin/notify-send", nil)
	doctor.StubProgramBaseName(t, "opencode-notify.exe")
	plugins := t.TempDir()
	reg := &aumid.FakeRegistry{Registered: true, DisplayName: aumid.AUMIDDisplayName}
	seedPluginContent(t, plugins, staleSeed())
	// Dead registry on purpose: no update-notice line must appear.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := server.URL
	server.Close()
	doctor.SwapRegistryURL(t, url)

	var out bytes.Buffer
	if code := doctor.Doctor(plugins, reg, &out); code != 1 {
		t.Fatalf("stale doctor exited %d, want 1:\n%s", code, out.String())
	}
	got := out.String()
	if !strings.Contains(got, "STALE (") {
		t.Fatalf("doctor must report STALE, got:\n%s", got)
	}
	if !strings.Contains(got, `Issues found: run "opencode-notify install" to fix.`) {
		t.Fatalf("stale doctor must keep the install-to-fix line, got:\n%s", got)
	}
}

func TestDoctorHealthyOutputUnchanged(t *testing.T) {
	doctor.StubNotifier(t, "/bin/notify-send", nil)
	plugins := t.TempDir()
	reg := &aumid.FakeRegistry{}
	doctor.SeedHealthyInstall(t, plugins, reg)
	// Dead registry on purpose: no update-notice line must appear.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := server.URL
	server.Close()
	doctor.SwapRegistryURL(t, url)

	var out bytes.Buffer
	if code := doctor.Doctor(plugins, reg, &out); code != 0 {
		t.Fatalf("healthy doctor exited %d:\n%s", code, out.String())
	}
	identity := "Identity"
	verdict := "Healthy: plugin installed and notifier ready."
	if runtime.GOOS == "windows" {
		identity = "AUMID registration"
		verdict = "Healthy: plugin installed and AUMID registered."
	}
	want := fmt.Sprintf(
		"Plugin file: OK (%s)\n%s: OK (registered (fake))\nNotifier: OK (/bin/notify-send)\n%s\n",
		config.PluginPath(plugins), identity, verdict)
	if out.String() != want {
		t.Fatalf("healthy doctor output changed!\n--- want ---\n%s\n--- got ---\n%s", want, out.String())
	}
}
