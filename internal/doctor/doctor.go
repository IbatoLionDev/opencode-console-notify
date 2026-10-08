// Package doctor is the parity gate: it verifies the SAME end state as
// install.ps1 (plugin file present + AUMID registered) regardless of
// which install path produced it.
package doctor

import (
	"fmt"
	"io"
	"runtime"

	"github.com/IbatoLionDev/opencode-console-notify/internal/aumid"
)

// Doctor prints the health report and returns exit code 0 when healthy,
// 1 when something is missing.
func Doctor(pluginsDir string, reg aumid.Registry, out io.Writer) int {
	report, err := CheckHealth(pluginsDir, reg)
	if err != nil {
		fmt.Fprintf(out, "Doctor failed: %s\n", err)
		return 1
	}

	fmt.Fprintf(out, "Plugin file: %s\n", report.PluginNote)
	fmt.Fprintf(out, "%s: %s\n", identityLabel(runtime.GOOS), report.AUMIDNote)
	fmt.Fprintf(out, "Notifier: %s\n", report.NotifierNote)
	code := printVerdict(out, runtime.GOOS, report.Healthy())
	maybePrintUpdateNotice(out)
	return code
}

// identityLabel names the identity half of the report for one OS.
func identityLabel(goos string) string {
	if goos == "windows" {
		return "AUMID registration"
	}
	return "Identity"
}

// printVerdict prints the healthy/issues verdict lines and returns the
// matching exit code.
func printVerdict(out io.Writer, goos string, healthy bool) int {
	if healthy {
		if goos == "windows" {
			fmt.Fprintf(out, "Healthy: plugin installed and AUMID registered.\n")
		} else {
			fmt.Fprintf(out, "Healthy: plugin installed and notifier ready.\n")
		}
		return 0
	}
	fmt.Fprintf(out, "Issues found: run \"opencode-notify install\" to fix.\n")
	return 1
}
