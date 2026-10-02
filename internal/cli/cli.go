// Package cli implements the opencode-notify commands (install, uninstall,
// doctor, test, upgrade) shared by the CLI entry point without admin rights.
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

// splitPluginsDirFlag extracts the value from the "--plugins-dir=DIR"
// and "-plugins-dir=DIR" equals forms.
func splitPluginsDirFlag(arg string) (string, bool) {
	if v, ok := strings.CutPrefix(arg, "--plugins-dir="); ok {
		return v, true
	}
	if v, ok := strings.CutPrefix(arg, "-plugins-dir="); ok {
		return v, true
	}
	return "", false
}

// commandAliases maps flag-style spellings to their command so every
// spelling behaves identically.
var commandAliases = map[string]string{
	"--help":    "help",
	"-h":        "help",
	"--version": "version",
	"-V":        "version",
}

// parser holds CLI parse state across arguments so the loop body stays flat.
type parser struct {
	command    string
	pluginsDir string
}

// takeValue consumes the value following a --plugins-dir flag.
func (p *parser) takeValue(args []string, i int) (int, error) {
	if i+1 >= len(args) {
		return 0, fmt.Errorf("flag %s needs a value", args[i])
	}
	p.pluginsDir = args[i+1]
	return 2, nil
}

// step processes one argument and returns how many args were consumed.
func (p *parser) step(args []string, i int) (int, error) {
	arg := args[i]
	if arg == "--plugins-dir" || arg == "-plugins-dir" {
		return p.takeValue(args, i)
	}
	if v, ok := splitPluginsDirFlag(arg); ok {
		p.pluginsDir = v
		return 1, nil
	}
	if alias, ok := commandAliases[arg]; ok {
		if p.command != "" {
			return 0, fmt.Errorf("unexpected argument %s", arg)
		}
		p.command = alias
		return 1, nil
	}
	if strings.HasPrefix(arg, "-") {
		return 0, fmt.Errorf("unknown flag %s", arg)
	}
	if p.command == "" {
		p.command = arg
		return 1, nil
	}
	return 0, fmt.Errorf("unexpected argument %s", arg)
}

// parseArgs extracts the subcommand and the --plugins-dir override.
// The flag is accepted before or after the command, as
// "--plugins-dir DIR" or "--plugins-dir=DIR" (single-dash form too).
func parseArgs(args []string) (command string, pluginsDir string, err error) {
	p := &parser{}
	for i := 0; i < len(args); {
		n, stepErr := p.step(args, i)
		if stepErr != nil {
			return "", "", stepErr
		}
		i += n
	}
	switch p.command {
	case "":
		return "", "", io.EOF // signal: print usage
	case "install", "uninstall", "doctor", "test", "upgrade", "version", "help":
		return p.command, p.pluginsDir, nil
	default:
		return "", "", fmt.Errorf("unknown command %q (want install|uninstall|doctor|test|upgrade|version)", p.command)
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
	default:
		fmt.Fprintf(stderr, "Error: unknown command %q\n\n%s", command, usageText)
		return 2
	}
}
