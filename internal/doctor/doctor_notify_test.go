package doctor

import (
	"errors"
	"testing"

	"github.com/IbatoLionDev/opencode-console-notify/internal/aumid"
)

// stubNotifier points the backend probe at a fixed result for one test.
// Tests must never depend on the machine PATH: a Linux box without
// notify-send (or Windows without PowerShell on PATH) would flip them.
func stubNotifier(t *testing.T, path string, err error) {
	t.Helper()
	old := NotifierLookPath
	NotifierLookPath = func(string) (string, error) { return path, err }
	t.Cleanup(func() { NotifierLookPath = old })
}

func TestNotifierPresentKeepsHealthWhole(t *testing.T) {
	stubNotifier(t, "/usr/bin/notify-send", nil)
	plugins := t.TempDir()
	reg := &aumid.FakeRegistry{Registered: true, DisplayName: aumid.AUMIDDisplayName}
	seedPluginFileForHealth(t, plugins, true, false)

	report, err := CheckHealth(plugins, reg)
	if err != nil {
		t.Fatalf("CheckHealth failed: %v", err)
	}
	if !report.NotifierOK || !report.Healthy() {
		t.Fatalf("present backend must keep health whole: %+v", report)
	}
}

func TestNotifierMissingFailsHealthWithHint(t *testing.T) {
	stubNotifier(t, "", errors.New("not found"))
	plugins := t.TempDir()
	reg := &aumid.FakeRegistry{Registered: true, DisplayName: aumid.AUMIDDisplayName}
	seedPluginFileForHealth(t, plugins, true, false)

	report, err := CheckHealth(plugins, reg)
	if err != nil {
		t.Fatalf("CheckHealth failed: %v", err)
	}
	if report.NotifierOK || report.Healthy() {
		t.Fatalf("missing backend must fail health: %+v", report)
	}
}
