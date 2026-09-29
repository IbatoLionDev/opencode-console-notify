// Package doctor is the parity gate: it verifies the SAME end state as
// install.ps1 (plugin file present + AUMID registered) regardless of
// which install path produced it.
package doctor

import (
	"fmt"
	"io"
	"os"

	"github.com/IbatoLionDev/opencode-console-notify/internal/aumid"
	"github.com/IbatoLionDev/opencode-console-notify/internal/config"
)

// HealthReport describes the install end state at one point in time.
type HealthReport struct {
	PluginsDir string
	PluginPath string
	PluginOK   bool
	PluginNote string
	AUMIDOK    bool
	AUMIDNote  string
}

// Healthy is true only when both halves of the end state hold.
func (r HealthReport) Healthy() bool { return r.PluginOK && r.AUMIDOK }

// CheckHealth inspects the plugin file and the AUMID registration.
// A missing plugin file or key is reported, never returned as an error;
// errors are reserved for genuine failures (unreadable directory,
// registry access failure).
func CheckHealth(pluginsDir string, reg aumid.Registry) (HealthReport, error) {
	report := HealthReport{
		PluginsDir: pluginsDir,
		PluginPath: config.PluginPath(pluginsDir),
	}

	info, err := os.Stat(report.PluginPath)
	switch {
	case err == nil:
		if info.IsDir() {
			report.PluginOK = false
			report.PluginNote = fmt.Sprintf("MISSING (%s is a directory, not a file)", report.PluginPath)
		} else if info.Size() == 0 {
			report.PluginOK = false
			report.PluginNote = fmt.Sprintf("MISSING (%s is empty)", report.PluginPath)
		} else {
			report.PluginOK = true
			report.PluginNote = fmt.Sprintf("OK (%s)", report.PluginPath)
		}
	case os.IsNotExist(err):
		report.PluginOK = false
		report.PluginNote = fmt.Sprintf("MISSING (no file at %s)", report.PluginPath)
	default:
		return report, fmt.Errorf("stat plugin file: %w", err)
	}

	registered, detail, err := reg.AUMIDStatus()
	if err != nil {
		return report, fmt.Errorf("check AUMID registration: %w", err)
	}
	report.AUMIDOK = registered
	if registered {
		report.AUMIDNote = "OK (" + detail + ")"
	} else {
		report.AUMIDNote = "MISSING (" + detail + ")"
	}

	return report, nil
}

// Doctor prints the health report and returns exit code 0 when healthy,
// 1 when something is missing.
func Doctor(pluginsDir string, reg aumid.Registry, out io.Writer) int {
	report, err := CheckHealth(pluginsDir, reg)
	if err != nil {
		fmt.Fprintf(out, "Doctor failed: %s\n", err)
		return 1
	}

	fmt.Fprintf(out, "Plugin file: %s\n", report.PluginNote)
	fmt.Fprintf(out, "AUMID registration: %s\n", report.AUMIDNote)
	if report.Healthy() {
		fmt.Fprintf(out, "Healthy: plugin installed and AUMID registered.\n")
		return 0
	}
	fmt.Fprintf(out, "Issues found: run \"opencode-notify install\" to fix.\n")
	return 1
}
