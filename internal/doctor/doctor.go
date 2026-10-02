// Package doctor is the parity gate: it verifies the SAME end state as
// install.ps1 (plugin file present + AUMID registered) regardless of
// which install path produced it.
package doctor

import (
	"fmt"
	"io"

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
	fmt.Fprintf(out, "AUMID registration: %s\n", report.AUMIDNote)
	code := printVerdict(out, report.Healthy())
	maybePrintUpdateNotice(out)
	return code
}

// printVerdict prints the healthy/issues verdict lines and returns the
// matching exit code.
func printVerdict(out io.Writer, healthy bool) int {
	if healthy {
		fmt.Fprintf(out, "Healthy: plugin installed and AUMID registered.\n")
		return 0
	}
	fmt.Fprintf(out, "Issues found: run \"opencode-notify install\" to fix.\n")
	return 1
}
