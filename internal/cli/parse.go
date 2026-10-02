// Parse holds CLI argument parsing: the subcommand and the
// --plugins-dir override. The flag is accepted before or after the
// command, as "--plugins-dir DIR" or "--plugins-dir=DIR" (single-dash
// form too).
package cli

import (
	"fmt"
	"io"
	"strings"
)

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
	rest       []string
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
// After the config command, every remaining argument (including flags)
// is collected as config args so `config --lang es` works for scripts.
func (p *parser) step(args []string, i int) (int, error) {
	arg := args[i]
	if p.command == "config" {
		if arg == "--plugins-dir" || arg == "-plugins-dir" {
			return p.takeValue(args, i)
		}
		if v, ok := splitPluginsDirFlag(arg); ok {
			p.pluginsDir = v
			return 1, nil
		}
		p.rest = append(p.rest, arg)
		return 1, nil
	}
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

// parseArgsFull extracts the subcommand, the --plugins-dir override,
// and the trailing config args (only for the config command).
func parseArgsFull(args []string) (command string, pluginsDir string, rest []string, err error) {
	p := &parser{}
	for i := 0; i < len(args); {
		n, stepErr := p.step(args, i)
		if stepErr != nil {
			return "", "", nil, stepErr
		}
		i += n
	}
	switch p.command {
	case "":
		return "", "", nil, io.EOF // signal: print usage
	case "install", "uninstall", "doctor", "test", "upgrade", "version", "help", "config":
		return p.command, p.pluginsDir, p.rest, nil
	default:
		return "", "", nil, fmt.Errorf("unknown command %q (want install|uninstall|doctor|test|upgrade|version|config)", p.command)
	}
}

// parseArgs extracts the subcommand and the --plugins-dir override.
func parseArgs(args []string) (command string, pluginsDir string, err error) {
	command, pluginsDir, _, err = parseArgsFull(args)
	return command, pluginsDir, err
}
