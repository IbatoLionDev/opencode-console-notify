// Command opencode-notify installs, checks, and exercises the
// opencode-console-notify Windows toast plugin without admin rights.
package main

import (
	"fmt"
	"io"
	"os"
	"strings"
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

// run executes the CLI; it returns the process exit code so tests can
// assert on it without spawning a subprocess.
func run(args []string, reg Registry, stdout, stderr io.Writer) int {
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

	pluginsDir, err := ResolvePluginsDir(override)
	if err != nil {
		fmt.Fprintf(stderr, "Error: %s\n", err)
		return 2
	}

	switch command {
	case "install":
		if err := Install(pluginsDir, reg, stdout); err != nil {
			fmt.Fprintf(stderr, "Error: %s\n", err)
			return 1
		}
		return 0
	case "uninstall":
		if err := Uninstall(pluginsDir, reg, stdout); err != nil {
			fmt.Fprintf(stderr, "Error: %s\n", err)
			return 1
		}
		return 0
	case "doctor":
		return Doctor(pluginsDir, reg, stdout)
	case "test":
		if err := SendTestToast(reg); err != nil {
			fmt.Fprintf(stderr, "Error: %s\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "Sent test notification via %s.\n", AUMID)
		return 0
	default:
		fmt.Fprintf(stderr, "Error: unknown command %q\n\n%s", command, usageText)
		return 2
	}
}

func main() {
	os.Exit(run(os.Args[1:], NewRegistry(), os.Stdout, os.Stderr))
}
