// Package tui implements the config command interactive menu.
// Stdlib only: line-based menu so it works on PowerShell 5.1 with no
// raw terminal mode. Every view prints its keys in a visible footer.
package tui

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/IbatoLionDev/opencode-console-notify/internal/config"
	"github.com/IbatoLionDev/opencode-console-notify/internal/i18n"
)

// Links shown in the info view. They mirror cli.go and doctor/notice.go.
const (
	NPMPage     = "https://www.npmjs.com/package/opencode-console-notify"
	RegistryURL = "https://registry.npmjs.org/opencode-console-notify/latest"
	RepoURL     = "https://github.com/IbatoLionDev/opencode-console-notify"
	ReleasesURL = "https://github.com/IbatoLionDev/opencode-console-notify/releases"
)

// AlertItem binds one toggle to its event, display name and trigger text.
type AlertItem struct {
	Key        string
	Event      string
	Name       string
	Trigger    string
	Enabled    bool
	SettingPtr *bool
}

// AlertItems returns the four default alerts in stable order.
func AlertItems(s *config.Settings, lang string) []AlertItem {
	return []AlertItem{
		{Key: "1", Event: "session.idle", Name: i18n.T(lang, "alert.sessionIdle"), Trigger: i18n.T(lang, "trigger.sessionIdle"), Enabled: s.Alerts.SessionIdle, SettingPtr: &s.Alerts.SessionIdle},
		{Key: "2", Event: "session.error", Name: i18n.T(lang, "alert.sessionError"), Trigger: i18n.T(lang, "trigger.sessionError"), Enabled: s.Alerts.SessionError, SettingPtr: &s.Alerts.SessionError},
		{Key: "3", Event: "permission.asked", Name: i18n.T(lang, "alert.permissionAsked"), Trigger: i18n.T(lang, "trigger.permissionAsked"), Enabled: s.Alerts.PermissionAsked, SettingPtr: &s.Alerts.PermissionAsked},
		{Key: "4", Event: "question.asked", Name: i18n.T(lang, "alert.questionAsked"), Trigger: i18n.T(lang, "trigger.questionAsked"), Enabled: s.Alerts.QuestionAsked, SettingPtr: &s.Alerts.QuestionAsked},
	}
}

func onOff(lang string, on bool) string {
	if on {
		return i18n.T(lang, "state.on")
	}
	return i18n.T(lang, "state.off")
}

// RenderMenu returns the main menu with its key footer.
func RenderMenu(lang string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", i18n.T(lang, "app.title"))
	fmt.Fprintf(&b, "1. %s (L)\n", i18n.T(lang, "menu.language"))
	fmt.Fprintf(&b, "2. %s (U)\n", i18n.T(lang, "menu.update"))
	fmt.Fprintf(&b, "3. %s (I)\n", i18n.T(lang, "menu.info"))
	fmt.Fprintf(&b, "4. %s (A)\n", i18n.T(lang, "menu.alerts"))
	fmt.Fprintf(&b, "5. %s (Q)\n", i18n.T(lang, "menu.exit"))
	fmt.Fprintf(&b, "%s\n", i18n.T(lang, "footer.menu"))
	return b.String()
}

// RenderLanguage returns the language view.
func RenderLanguage(lang string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", i18n.T(lang, "lang.title"))
	fmt.Fprintf(&b, "1. English%s\n", marker(lang == "en", lang))
	fmt.Fprintf(&b, "2. Español%s\n", marker(lang == "es", lang))
	fmt.Fprintf(&b, "[1/2] %s | %s\n", i18n.T(lang, "lang.current")+": "+lang, i18n.T(lang, "footer.back"))
	return b.String()
}

func marker(active bool, lang string) string {
	if active {
		return " [" + i18n.T(lang, "lang.current") + "]"
	}
	return ""
}

// RenderAlerts returns the alerts view with toggles and triggers.
func RenderAlerts(s *config.Settings, lang string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", i18n.T(lang, "alerts.title"))
	for _, a := range AlertItems(s, lang) {
		fmt.Fprintf(&b, "%s. %s [%s]\n   %s\n", a.Key, a.Name, onOff(lang, a.Enabled), a.Trigger)
	}
	fmt.Fprintf(&b, "%s\n", i18n.T(lang, "footer.toggle"))
	return b.String()
}

// RenderInfo returns the commands plus documentation links.
func RenderInfo(lang string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", i18n.T(lang, "menu.info"))
	fmt.Fprintf(&b, "install, uninstall, doctor, test, upgrade, version, config\n")
	fmt.Fprintf(&b, "npm: %s\n", NPMPage)
	fmt.Fprintf(&b, "repo: %s\n", RepoURL)
	fmt.Fprintf(&b, "releases: %s\n", ReleasesURL)
	fmt.Fprintf(&b, "%s\n", i18n.T(lang, "footer.back"))
	return b.String()
}

func readKey(r *bufio.Reader) string {
	line, err := r.ReadString('\n')
	if err != nil && len(line) == 0 {
		return "q"
	}
	return strings.ToLower(strings.TrimSpace(line))
}

// Run opens the interactive loop. Upgrade runs the existing upgrade flow
// for the current distribution and returns its exit code. Run returns the
// process exit code for the config command.
func Run(pluginsDir string, stdin io.Reader, stdout io.Writer, upgrade func(io.Writer) int) int {
	s, err := config.Load(pluginsDir)
	if err != nil {
		fmt.Fprintf(stdout, "Error: %s\n", err)
		return 1
	}
	lang := i18n.Normalize(s.Lang)
	r := bufio.NewReader(stdin)
	view := "menu"
	for {
		switch view {
		case "menu":
			fmt.Fprint(stdout, RenderMenu(lang))
			switch readKey(r) {
			case "1", "l", "language", "idioma":
				view = "language"
			case "2", "u", "update", "actualizar":
				view = "update"
			case "3", "i", "info":
				view = "info"
			case "4", "a", "alerts", "alertas":
				view = "alerts"
			case "5", "q", "quit", "exit", "salir", "esc":
				return 0
			}
		case "language":
			fmt.Fprint(stdout, RenderLanguage(lang))
			switch readKey(r) {
			case "1", "en":
				lang = "en"
				s.Lang = lang
				if err := config.Save(pluginsDir, s); err != nil {
					fmt.Fprintf(stdout, "Error: %s\n", err)
					return 1
				}
			case "2", "es":
				lang = "es"
				s.Lang = lang
				if err := config.Save(pluginsDir, s); err != nil {
					fmt.Fprintf(stdout, "Error: %s\n", err)
					return 1
				}
			case "q", "esc", "b", "back", "volver":
				view = "menu"
			}
		case "alerts":
			fmt.Fprint(stdout, RenderAlerts(&s, lang))
			key := readKey(r)
			if key == "q" || key == "esc" || key == "b" || key == "back" || key == "volver" {
				view = "menu"
				continue
			}
			// Space-prefixed numbers (" 1") also toggle: TrimSpace already
			// removed the space, so the digit alone is enough.
			for _, a := range AlertItems(&s, lang) {
				if key == a.Key {
					*a.SettingPtr = !*a.SettingPtr
					if err := config.Save(pluginsDir, s); err != nil {
						fmt.Fprintf(stdout, "Error: %s\n", err)
						return 1
					}
				}
			}
		case "info":
			fmt.Fprint(stdout, RenderInfo(lang))
			readKey(r)
			view = "menu"
		case "update":
			if upgrade != nil {
				return upgrade(stdout)
			}
			return 0
		}
	}
}
