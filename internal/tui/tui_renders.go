// Line-mode renders: pure string builders for the config views.
// Interaction (session, loop) lives in tui.go.
package tui

import (
	"fmt"
	"strconv"
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

// AllAlertItems returns defaults followed by customs tagged as custom,
// numbered from 1 in stable order. Customs get positional keys ("5", …)
// so both TUI modes toggle them exactly like defaults.
func AllAlertItems(s *config.Settings, lang string) []AlertItem {
	items := AlertItems(s, lang)
	for i := range s.CustomAlerts {
		c := &s.CustomAlerts[i]
		trigger := c.Event
		if c.Body != "" {
			trigger += " — " + c.Body
		}
		items = append(items, AlertItem{
			Key:        strconv.Itoa(len(items) + 1),
			Event:      c.Event,
			Name:       c.Title + " [" + i18n.T(lang, "custom.tag") + "]",
			Trigger:    trigger,
			Enabled:    c.Enabled,
			SettingPtr: &c.Enabled,
		})
	}
	return items
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
	fmt.Fprintf(&b, "5. %s (C)\n", i18n.T(lang, "menu.customs"))
	fmt.Fprintf(&b, "6. %s (Q)\n", i18n.T(lang, "menu.exit"))
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

// RenderAlerts returns the alerts view with toggles and triggers,
// defaults first then customs tagged as custom.
func RenderAlerts(s *config.Settings, lang string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", i18n.T(lang, "alerts.title"))
	for _, a := range AllAlertItems(s, lang) {
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
