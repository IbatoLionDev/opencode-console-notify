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
			fmt.Fprintf(out, errorFormat, err)
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
		fmt.Fprintf(stdout, errorFormat, err)
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
		fmt.Fprintf(stdout, errorFormat, err)
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
		fmt.Fprintf(stdout, errorFormat, err)
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
		fmt.Fprintf(stdout, errorFormat, err)
		return 1
	}
	s.Lang = lang
	if err := config.Save(pluginsDir, s); err != nil {
		fmt.Fprintf(stdout, errorFormat, err)
		return 1
	}
	fmt.Fprintf(stdout, "Language: %s\n", lang)
	return 0
}

func printConfigInfo(pluginsDir string, stdout io.Writer) int {
	s, err := config.Load(pluginsDir)
	if err != nil {
		fmt.Fprintf(stdout, errorFormat, err)
		return 1
	}
	fmt.Fprint(stdout, tui.RenderInfo(i18n.Normalize(s.Lang)))
	return 0
}

func takeConfigValue(configArgs []string, i int, stdout io.Writer) (string, bool) {
	if i+1 >= len(configArgs) {
		fmt.Fprintf(stdout, "Error: flag %s needs a value\n", configArgs[i])
		return "", false
	}
	return configArgs[i+1], true
}

// dispatchValueFlag handles --lang/--toggle in both `--flag value` and
// `--flag=value` forms. The second return reports whether arg was a value
// flag at all.
func dispatchValueFlag(pluginsDir string, configArgs []string, stdout io.Writer) (int, bool) {
	arg := configArgs[0]
	switch {
	case arg == "--lang" || arg == "-lang":
		value, ok := takeConfigValue(configArgs, 0, stdout)
		if !ok {
			return 2, true
		}
		return applyLang(pluginsDir, value, stdout), true
	case strings.HasPrefix(arg, "--lang="):
		return applyLang(pluginsDir, strings.TrimPrefix(arg, "--lang="), stdout), true
	case arg == "--toggle" || arg == "-toggle":
		value, ok := takeConfigValue(configArgs, 0, stdout)
		if !ok {
			return 2, true
		}
		return applyToggle(pluginsDir, value, stdout), true
	case strings.HasPrefix(arg, "--toggle="):
		return applyToggle(pluginsDir, strings.TrimPrefix(arg, "--toggle="), stdout), true
	}
	return 0, false
}

func printConfigEvents(pluginsDir string, stdout io.Writer) int {
	s, err := config.Load(pluginsDir)
	if err != nil {
		fmt.Fprintf(stdout, errorFormat, err)
		return 1
	}
	lang := i18n.Normalize(s.Lang)
	for i, e := range config.KnownEvents() {
		desc := e.DescEn
		if lang == "es" {
			desc = e.DescEs
		}
		mark := ""
		if e.Noisy {
			mark = " [!]"
		}
		fmt.Fprintf(stdout, "%d. %s: %s%s\n", i+1, e.Type, desc, mark)
	}
	return 0
}

func printConfigCustoms(pluginsDir string, stdout io.Writer) int {
	s, err := config.Load(pluginsDir)
	if err != nil {
		fmt.Fprintf(stdout, errorFormat, err)
		return 1
	}
	if len(s.CustomAlerts) == 0 {
		fmt.Fprintf(stdout, "%s\n", i18n.T(s.Lang, "customs.empty"))
		return 0
	}
	lang := i18n.Normalize(s.Lang)
	for _, c := range s.CustomAlerts {
		state := "state.on"
		if !c.Enabled {
			state = "state.off"
		}
		fmt.Fprintf(stdout, "%s: %s [%s] [%s]\n", c.ID, c.Title, i18n.T(lang, "custom.tag"), i18n.T(lang, state))
	}
	return 0
}

func dispatchConfigFlag(pluginsDir string, configArgs []string, stdout io.Writer) int {
	arg := configArgs[0]
	switch {
	case arg == "--list-alerts" || arg == "-list-alerts":
		return printAlerts(pluginsDir, stdout)
	case arg == "--list-events" || arg == "-list-events":
		return printConfigEvents(pluginsDir, stdout)
	case arg == "--list-customs" || arg == "-list-customs":
		return printConfigCustoms(pluginsDir, stdout)
	case arg == "--info" || arg == "-info":
		return printConfigInfo(pluginsDir, stdout)
	}
	if code, handled := dispatchValueFlag(pluginsDir, configArgs, stdout); handled {
		return code
	}
	fmt.Fprintf(stdout, "Error: unknown config flag %q (want --lang|--toggle|--list-alerts|--list-events|--list-customs|--info)\n", arg)
	return 2
}

// runConfig executes the config command. With no args it opens the TUI;
// flags keep scripts non-interactive.
func runConfig(pluginsDir string, configArgs []string, reg aumid.Registry, stdout io.Writer) int {
	if len(configArgs) == 0 {
		return tui.RunSmart(pluginsDir, os.Stdin, stdout, upgradeFromEmbedded(pluginsDir, reg, stdout))
	}
	return dispatchConfigFlag(pluginsDir, configArgs, stdout)
}
