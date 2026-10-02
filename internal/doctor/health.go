// Health holds the install end-state report and its inspection: the
// plugin file plus the AUMID registration. A missing file or key is
// reported, never returned as an error; errors are reserved for genuine
// failures (unreadable directory, registry access failure).
package doctor

import (
	"fmt"
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
