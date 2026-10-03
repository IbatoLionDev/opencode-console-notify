// Fullscreen menu and info views. Alerts and language live in their own
// files; the loop and wiring live in fullscreen.go.
package tui

import (
	"fmt"

	"github.com/IbatoLionDev/opencode-console-notify/internal/i18n"
	"github.com/IbatoLionDev/opencode-console-notify/internal/screen"
)

func menuScreenItems(lang string) []screen.Item {
	return []screen.Item{
		{Label: i18n.T(lang, "menu.language")},
		{Label: i18n.T(lang, "menu.update")},
		{Label: i18n.T(lang, "menu.info")},
		{Label: i18n.T(lang, "menu.alerts")},
		{Label: i18n.T(lang, "menu.exit")},
	}
}

func menuTarget(sel int) string {
	switch sel {
	case 0:
		return "language"
	case 1:
		return "update"
	case 2:
		return "info"
	case 3:
		return "alerts"
	}
	return "exit"
}

func menuShortcut(r rune) string {
	switch r {
	case '1', 'l', 'L':
		return "language"
	case '2', 'u', 'U':
		return "update"
	case '3', 'i', 'I':
		return "info"
	case '4', 'a', 'A':
		return "alerts"
	case '5':
		return "exit"
	}
	return ""
}

func (l *screenLoop) runMenu(sel *int) string {
	items := menuScreenItems(l.lang)
	*sel = screen.ClampSelection(len(items), *sel)
	frame := screen.Frame{
		Title:    i18n.T(l.lang, "app.title"),
		Items:    items,
		Selected: *sel,
		Footer:   i18n.T(l.lang, "footer.menu"),
		Height:   screen.FrameHeight(l.height),
	}
	fmt.Fprint(l.out, screen.Render(frame, l.width))
	k, err := l.readKey()
	if err != nil {
		return "abort"
	}
	if k.Key == screen.KeyRune {
		if target := menuShortcut(k.Rune); target != "" {
			return target
		}
		return "menu"
	}
	switch k.Key {
	case screen.KeyUp:
		*sel = screen.MoveSteps(len(items), *sel, -1)
	case screen.KeyDown:
		*sel = screen.MoveSteps(len(items), *sel, 1)
	case screen.KeyEnter:
		return menuTarget(*sel)
	case screen.KeyQuit, screen.KeyEsc:
		return "exit"
	}
	return "menu"
}

func (l *screenLoop) runInfo() string {
	frame := screen.Frame{
		Title:    i18n.T(l.lang, "menu.info"),
		Items:    infoScreenItems(),
		Selected: -1,
		Footer:   i18n.T(l.lang, "footer.back"),
		Height:   screen.FrameHeight(l.height),
	}
	fmt.Fprint(l.out, screen.Render(frame, l.width))
	if _, err := l.readKey(); err != nil {
		return "abort"
	}
	return "menu"
}

func infoScreenItems() []screen.Item {
	return []screen.Item{
		{Label: "install, uninstall, doctor, test, upgrade, version, config"},
		{Label: "npm: " + NPMPage},
		{Label: "repo: " + RepoURL},
		{Label: "releases: " + ReleasesURL},
	}
}
