package doctor

import (
	"errors"
	"testing"

	"github.com/IbatoLionDev/opencode-console-notify/internal/aumid"
)

func stubLookPath(path string, err error) func(string) (string, error) {
	return func(string) (string, error) { return path, err }
}

func TestNotifierPresentKeepsHealthWhole(t *testing.T) {
	plugins := t.TempDir()
	reg := &aumid.FakeRegistry{Registered: true, DisplayName: aumid.AUMIDDisplayName}
	seedPluginFileForHealth(t, plugins, true, false)

	report, err := CheckHealthWithLookPath(plugins, reg, stubLookPath("/usr/bin/notify-send", nil))
	if err != nil {
		t.Fatalf("CheckHealth failed: %v", err)
	}
	if !report.NotifierOK || !report.Healthy() {
		t.Fatalf("present backend must keep health whole: %+v", report)
	}
}

func TestNotifierMissingFailsHealthWithHint(t *testing.T) {
	plugins := t.TempDir()
	reg := &aumid.FakeRegistry{Registered: true, DisplayName: aumid.AUMIDDisplayName}
	seedPluginFileForHealth(t, plugins, true, false)

	report, err := CheckHealthWithLookPath(plugins, reg, stubLookPath("", errors.New("not found")))
	if err != nil {
		t.Fatalf("CheckHealth failed: %v", err)
	}
	if report.NotifierOK || report.Healthy() {
		t.Fatalf("missing backend must fail health: %+v", report)
	}
}
