// Package cli implements the opencode-notify commands (install, uninstall,
// doctor, test) shared by the CLI entry point without admin rights.
package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/IbatoLionDev/opencode-console-notify/internal/aumid"
	"github.com/IbatoLionDev/opencode-console-notify/internal/config"
	"github.com/IbatoLionDev/opencode-console-notify/internal/doctor"
	"github.com/IbatoLionDev/opencode-console-notify/internal/installer"
	"github.com/IbatoLionDev/opencode-console-notify/internal/notifier"
)

const usageText = `Usage: opencode-notify [--plugins-dir DIR] <command>

Commands:
  install     Copy the embedded plugin into the plugins directory and
              register the OpenCode.Notifier AUMID (HKCU, no admin).
  uninstall   Remove the plugin file and the AUMID key only.
  doctor      Verify the install end state (plugin file + AUMID).
              Exit 0 when healthy, 1 when something is missing.
  test        Send a real Windows toast notification.

Options:
  --plugins-dir DIR   Override the plugins directory.
                      Default: <HOME>/.config/opencode/plugins
`

// parseArgs extracts the subcommand and the --plugins-dir override.
// The flag is accepted before or after the command, as
// "--plugins-dir DIR" or "--plugins-dir=DIR" (single-dash form too).
func parseArgs(args []string) (command string, pluginsDir string, err error) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--plugins-dir" || arg == "-plugins-dir" {
			if i+1 >= len(args) {
				return "", "", fmt.Errorf("flag %s needs a value", arg)
			}
			i++
			pluginsDir = args[i]
			continue
		}
		if v, ok := strings.CutPrefix(arg, "--plugins-dir="); ok {
			pluginsDir = v
			continue
		}
		if v, ok := strings.CutPrefix(arg, "-plugins-dir="); ok {
			pluginsDir = v
			continue
		}
		if strings.HasPrefix(arg, "-") {
			return "", "", fmt.Errorf("unknown flag %s", arg)
		}
		if command == "" {
			command = arg
			continue
		}
		return "", "", fmt.Errorf("unexpected argument %s", arg)
	}
	switch command {
	case "":
		return "", "", io.EOF // signal: print usage
	case "install", "uninstall", "doctor", "test", "help", "--help", "-h":
		return command, pluginsDir, nil
	default:
		return "", "", fmt.Errorf("unknown command %q (want install|uninstall|doctor|test)", command)
	}
}

// Run executes the CLI; it returns the process exit code so tests can
// assert on it without spawning a subprocess.
func Run(args []string, reg aumid.Registry, stdout, stderr io.Writer) int {
	command, override, err := parseArgs(args)
	if err == io.EOF {
		fmt.Fprint(stdout, usageText)
		return 2
	}
	if err != nil {
		fmt.Fprintf(stderr, "Error: %s\n\n%s", err, usageText)
		return 2
	}
	if command == "help" || command == "--help" || command == "-h" {
		fmt.Fprint(stdout, usageText)
		return 0
	}

	pluginsDir, err := config.ResolvePluginsDir(override)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %s\n", err)
		return 2
	}

	switch command {
	case "install":
		if err := installer.Install(pluginsDir, reg, stdout); err != nil {
			fmt.Fprintf(stderr, "Error: %s\n", err)
			return 1
		}
		return 0
	case "uninstall":
		if err := installer.Uninstall(pluginsDir, reg, stdout); err != nil {
			fmt.Fprintf(stderr, "Error: %s\n", err)
			return 1
		}
		return 0
	case "doctor":
		return doctor.Doctor(pluginsDir, reg, stdout)
	case "test":
		if err := notifier.SendTestToast(reg); err != nil {
			fmt.Fprintf(stderr, "Error: %s\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Sent test notification via %s.\n", aumid.AUMID)
		return 0
	default:
		fmt.Fprintf(stderr, "Error: unknown command %q\n\n%s", command, usageText)
		return 2
	}
}
