package doctor_test

import (
	"errors"
	"testing"

	"github.com/IbatoLionDev/opencode-console-notify/internal/aumid"
	"github.com/IbatoLionDev/opencode-console-notify/internal/doctor"
)

func TestNotifierPresentKeepsHealthWhole(t *testing.T) {
	doctor.StubNotifier(t, "/usr/bin/notify-send", nil)
	plugins := t.TempDir()
	reg := &aumid.FakeRegistry{Registered: true, DisplayName: aumid.AUMIDDisplayName}
	doctor.SeedPluginFileForHealth(t, plugins, true, false)

	report, err := doctor.CheckHealth(plugins, reg)
	if err != nil {
		t.Fatalf("CheckHealth failed: %v", err)
	}
	if !report.NotifierOK || !report.Healthy() {
		t.Fatalf("present backend must keep health whole: %+v", report)
	}
}

func TestNotifierMissingFailsHealthWithHint(t *testing.T) {
	doctor.StubNotifier(t, "", errors.New("not found"))
	plugins := t.TempDir()
	reg := &aumid.FakeRegistry{Registered: true, DisplayName: aumid.AUMIDDisplayName}
	doctor.SeedPluginFileForHealth(t, plugins, true, false)

	report, err := doctor.CheckHealth(plugins, reg)
	if err != nil {
		t.Fatalf("CheckHealth failed: %v", err)
	}
	if report.NotifierOK || report.Healthy() {
		t.Fatalf("missing backend must fail health: %+v", report)
	}
}
