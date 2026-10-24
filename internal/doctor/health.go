// Health holds the install end-state report and its inspection: the
// plugin file, the AUMID registration, and a working notifier backend.
// A missing piece is reported, never returned as an error; errors are
// reserved for genuine failures (unreadable directory, registry access
// failure).
package doctor

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/IbatoLionDev/opencode-console-notify/internal/aumid"
	"github.com/IbatoLionDev/opencode-console-notify/internal/config"
	"github.com/IbatoLionDev/opencode-console-notify/plugin"
)

// HealthReport describes the install end state at one point in time.
type HealthReport struct {
	PluginsDir   string
	PluginPath   string
	PluginOK     bool
	PluginNote   string
	AUMIDOK      bool
	AUMIDNote    string
	NotifierOK   bool
	NotifierNote string
}

// Healthy is true only when every half of the end state holds.
func (r HealthReport) Healthy() bool { return r.PluginOK && r.AUMIDOK && r.NotifierOK }

// NotifierLookPath probes the toast backend binary. It is a variable
// (not exec.LookPath inline) so tests stay hermetic on every OS:
// unit tests point it at a stub, production leaves the default.
var NotifierLookPath = exec.LookPath

// ProgramBaseName resolves the running binary's own base name for the
// STALE install hint (filepath.Base(os.Args[0]): opencode-notify.exe on
// Windows, the linux/darwin binary name elsewhere). It is a variable
// (not filepath.Base inline) so tests stay hermetic: unit tests pin it,
// production leaves the default.
var ProgramBaseName = func() string { return filepath.Base(os.Args[0]) }

// CheckHealth inspects the plugin file, the AUMID registration, and the
// notifier backend available on this OS.
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
			installed, err := os.ReadFile(report.PluginPath)
			if err != nil {
				return report, fmt.Errorf("read plugin file: %w", err)
			}
			if bytes.Equal(installed, plugin.Source) {
				report.PluginOK = true
				report.PluginNote = fmt.Sprintf("OK (%s)", report.PluginPath)
			} else {
				report.PluginOK = false
				report.PluginNote = fmt.Sprintf(
					"STALE (%s differs from this build; run \"%s install\" to rewrite it)",
					report.PluginPath, ProgramBaseName())
			}
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

	report.NotifierOK, report.NotifierNote = checkNotifier(runtime.GOOS, NotifierLookPath)

	return report, nil
}

// checkNotifier verifies the toast backend binary for one OS exists on
// PATH: powershell.exe on Windows, notify-send on Linux. Anything else
// is an unsupported platform, reported rather than errored.
func checkNotifier(goos string, lookPath func(string) (string, error)) (bool, string) {
	binary := ""
	switch goos {
	case "windows":
		binary = "powershell.exe"
	case "linux":
		binary = "notify-send"
	case "darwin":
		binary = "osascript"
	default:
		return false, "MISSING (notifications not supported on " + goos + ")"
	}
	path, err := lookPath(binary)
	if err != nil {
		return false, "MISSING (" + binary + " not found on PATH" + missingHint(goos) + ")"
	}
	return true, "OK (" + path + ")"
}

// missingHint points at the package that provides the Linux backend;
// Windows needs no hint (PowerShell ships with the OS).
func missingHint(goos string) string {
	if goos == "linux" {
		return "; install libnotify-bin"
	}
	return ""
}
