// Package cli implements the opencode-notify commands (install, uninstall,
// doctor, test, upgrade, config) shared by the CLI entry point without admin rights.
package cli

import (
	"fmt"
	"io"

	"github.com/IbatoLionDev/opencode-console-notify/internal/aumid"
	"github.com/IbatoLionDev/opencode-console-notify/internal/config"
	"github.com/IbatoLionDev/opencode-console-notify/internal/doctor"
	"github.com/IbatoLionDev/opencode-console-notify/internal/installer"
	"github.com/IbatoLionDev/opencode-console-notify/internal/notifier"
	"github.com/IbatoLionDev/opencode-console-notify/internal/version"
)

const usageText = `Usage: opencode-notify [--plugins-dir DIR] <command>

Commands:
  install     Copy the embedded plugin into the plugins directory and
              register the OpenCode.Notifier AUMID (HKCU, no admin).
  uninstall   Remove the plugin file and the AUMID key only.
  doctor      Verify the install end state (plugin file + AUMID).
              Exit 0 when healthy, 1 when something is missing.
  test        Send a real Windows toast notification.
  upgrade     Reinstall the plugin from the embedded copy and verify.
              Newer binaries come from the GitHub Releases page.
  config      Open the config TUI (language, update, info, alerts).
              Flags for scripts: --lang en|es, --toggle KEY=on|off,
              --list-alerts, --info.
  version     Print the embedded binary version.

Options:
  --plugins-dir DIR   Override the plugins directory.
                      Default: <HOME>/.config/opencode/plugins
  --version, -V       Print the version and exit (same as version).
`

// releasesURL is where newer opencode-notify.exe binaries are published.
// The embedded plugin copy is build-pinned, so upgrade says so honestly.
const releasesURL = "https://github.com/IbatoLionDev/opencode-console-notify/releases"

const errorFormat = "Error: %s\n"

// Run executes the CLI; it returns the process exit code so tests can
// assert on it without spawning a subprocess.
func Run(args []string, reg aumid.Registry, stdout, stderr io.Writer) int {
	command, override, configArgs, err := parseArgsFull(args)
	if err == io.EOF {
		fmt.Fprint(stdout, usageText)
		return 2
	}
	if err != nil {
		fmt.Fprintf(stderr, "Error: %s\n\n%s", err, usageText)
		return 2
	}
	if command == "help" {
		fmt.Fprint(stdout, usageText)
		return 0
	}

	pluginsDir, err := config.ResolvePluginsDir(override)
	if err != nil {
		fmt.Fprintf(stderr, errorFormat, err)
		return 2
	}

	switch command {
	case "install":
		if err := installer.Install(pluginsDir, reg, stdout); err != nil {
			fmt.Fprintf(stderr, errorFormat, err)
			return 1
		}
		return 0
	case "uninstall":
		if err := installer.Uninstall(pluginsDir, reg, stdout); err != nil {
			fmt.Fprintf(stderr, errorFormat, err)
			return 1
		}
		return 0
	case "doctor":
		return doctor.Doctor(pluginsDir, reg, stdout)
	case "test":
		if err := notifier.SendTestToast(reg); err != nil {
			fmt.Fprintf(stderr, errorFormat, err)
			return 1
		}
		fmt.Fprintf(stdout, "Sent test notification via %s.\n", aumid.AUMID)
		return 0
	case "upgrade":
		fmt.Fprintf(stdout, "Upgrading with embedded copy (version %s)...\n", version.Version)
		if err := installer.Install(pluginsDir, reg, stdout); err != nil {
			fmt.Fprintf(stderr, errorFormat, err)
			return 1
		}
		code := doctor.Doctor(pluginsDir, reg, stdout)
		fmt.Fprintf(stdout, "Upgrade complete: version %s.\n", version.Version)
		fmt.Fprintf(stdout, "Note: this binary carries a build-pinned copy; newer binaries come from the GitHub Releases page: %s\n", releasesURL)
		return code
	case "version":
		fmt.Fprintf(stdout, "opencode-notify %s\n", version.Version)
		return 0
	case "config":
		return runConfig(pluginsDir, configArgs, reg, stdout)
	default:
		fmt.Fprintf(stderr, "Error: unknown command %q\n\n%s", command, usageText)
		return 2
	}
}
