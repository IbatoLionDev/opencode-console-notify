// Config command: non-interactive flags for scripts plus the
// interactive TUI. Update reuses the embedded reinstall flow.
package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/IbatoLionDev/opencode-console-notify/internal/aumid"
	"github.com/IbatoLionDev/opencode-console-notify/internal/config"
	"github.com/IbatoLionDev/opencode-console-notify/internal/doctor"
	"github.com/IbatoLionDev/opencode-console-notify/internal/i18n"
	"github.com/IbatoLionDev/opencode-console-notify/internal/installer"
	"github.com/IbatoLionDev/opencode-console-notify/internal/tui"
	"github.com/IbatoLionDev/opencode-console-notify/internal/version"
)

func upgradeFromEmbedded(pluginsDir string, reg aumid.Registry, stdout io.Writer) func(io.Writer) int {
	return func(out io.Writer) int {
		fmt.Fprintf(out, "Upgrading with embedded copy (version %s)...\n", version.Version)
		if err := installer.Install(pluginsDir, reg, out); err != nil {
			fmt.Fprintf(out, "Error: %s\n", err)
			return 1
		}
		code := doctor.Doctor(pluginsDir, reg, out)
		fmt.Fprintf(out, "Upgrade complete: version %s.\n", version.Version)
		fmt.Fprintf(out, "Note: this binary carries a build-pinned copy; newer binaries come from the GitHub Releases page: %s\n", releasesURL)
		_ = stdout
		return code
	}
}

func printAlerts(pluginsDir string, stdout io.Writer) int {
	s, err := config.Load(pluginsDir)
	if err != nil {
		fmt.Fprintf(stdout, "Error: %s\n", err)
		return 1
	}
	lang := i18n.Normalize(s.Lang)
	fmt.Fprint(stdout, tui.RenderAlerts(&s, lang))
	return 0
}

func applyToggle(pluginsDir, spec string, stdout io.Writer) int {
	key, value, ok := strings.Cut(spec, "=")
	if !ok {
		fmt.Fprintf(stdout, "Error: --toggle needs KEY=on|off (got %q)\n", spec)
		return 2
	}
	on := value == "on" || value == "true" || value == "1"
	off := value == "off" || value == "false" || value == "0"
	if !on && !off {
		fmt.Fprintf(stdout, "Error: --toggle value must be on or off (got %q)\n", value)
		return 2
	}
	s, err := config.Load(pluginsDir)
	if err != nil {
		fmt.Fprintf(stdout, "Error: %s\n", err)
		return 1
	}
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "sessionidle", "session.idle", "idle":
		s.Alerts.SessionIdle = on
	case "sessionerror", "session.error", "error":
		s.Alerts.SessionError = on
	case "permissionasked", "permission.asked", "permission", "perm":
		s.Alerts.PermissionAsked = on
	case "questionasked", "question.asked", "question":
		s.Alerts.QuestionAsked = on
	default:
		fmt.Fprintf(stdout, "Error: unknown alert %q (want sessionIdle|sessionError|permissionAsked|questionAsked)\n", key)
		return 2
	}
	if err := config.Save(pluginsDir, s); err != nil {
		fmt.Fprintf(stdout, "Error: %s\n", err)
		return 1
	}
	return printAlerts(pluginsDir, stdout)
}

func applyLang(pluginsDir, lang string, stdout io.Writer) int {
	lang = strings.ToLower(strings.TrimSpace(lang))
	if lang != "en" && lang != "es" {
		fmt.Fprintf(stdout, "Error: --lang must be en or es (got %q)\n", lang)
		return 2
	}
	s, err := config.Load(pluginsDir)
	if err != nil {
		fmt.Fprintf(stdout, "Error: %s\n", err)
		return 1
	}
	s.Lang = lang
	if err := config.Save(pluginsDir, s); err != nil {
		fmt.Fprintf(stdout, "Error: %s\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "Language: %s\n", lang)
	return 0
}

// runConfig executes the config command. With no args it opens the TUI;
// flags keep scripts non-interactive.
func runConfig(pluginsDir string, configArgs []string, reg aumid.Registry, stdout io.Writer) int {
	for i := 0; i < len(configArgs); i++ {
		arg := configArgs[i]
		switch {
		case arg == "--list-alerts" || arg == "-list-alerts":
			return printAlerts(pluginsDir, stdout)
		case arg == "--info" || arg == "-info":
			s, err := config.Load(pluginsDir)
			if err != nil {
				fmt.Fprintf(stdout, "Error: %s\n", err)
				return 1
			}
			fmt.Fprint(stdout, tui.RenderInfo(i18n.Normalize(s.Lang)))
			return 0
		case arg == "--lang" || arg == "-lang":
			if i+1 >= len(configArgs) {
				fmt.Fprintf(stdout, "Error: flag %s needs a value\n", arg)
				return 2
			}
			return applyLang(pluginsDir, configArgs[i+1], stdout)
		case strings.HasPrefix(arg, "--lang="):
			return applyLang(pluginsDir, strings.TrimPrefix(arg, "--lang="), stdout)
		case arg == "--toggle" || arg == "-toggle":
			if i+1 >= len(configArgs) {
				fmt.Fprintf(stdout, "Error: flag %s needs a value\n", arg)
				return 2
			}
			return applyToggle(pluginsDir, configArgs[i+1], stdout)
		case strings.HasPrefix(arg, "--toggle="):
			return applyToggle(pluginsDir, strings.TrimPrefix(arg, "--toggle="), stdout)
		default:
			fmt.Fprintf(stdout, "Error: unknown config flag %q (want --lang|--toggle|--list-alerts|--info)\n", arg)
			return 2
		}
	}
	return tui.Run(pluginsDir, os.Stdin, stdout, upgradeFromEmbedded(pluginsDir, reg, stdout))
}
