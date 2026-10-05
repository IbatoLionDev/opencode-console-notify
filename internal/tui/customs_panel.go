// Fullscreen customs panel: list, toggle and delete over user
// notifications. Add/edit flows with text input live in customs_editor.go.
package tui

import (
	"fmt"

	"github.com/IbatoLionDev/opencode-console-notify/internal/config"
	"github.com/IbatoLionDev/opencode-console-notify/internal/i18n"
	"github.com/IbatoLionDev/opencode-console-notify/internal/screen"
)

// customsPanelItems renders one row per custom: title tagged custom with
// on/off state, event plus body as the detail line.
func customsPanelItems(s *config.Settings, lang string) []screen.Item {
	items := make([]screen.Item, 0, len(s.CustomAlerts))
	for _, c := range s.CustomAlerts {
		label := c.Title + " [" + i18n.T(lang, "custom.tag") + "]"
		detail := c.Event
		if c.Body != "" {
			detail += " — " + c.Body
		}
		items = append(items, screen.Item{
			Label:  label,
			Detail: detail,
			State:  i18n.T(lang, boolStateKey(c.Enabled)),
		})
	}
	if len(items) == 0 {
		items = append(items, screen.Item{Label: i18n.T(lang, "customs.empty")})
	}
	return items
}

func (l *screenLoop) runCustoms(sel *int) string {
	customs := l.settings.CustomAlerts
	*sel = screen.ClampSelection(len(customs), *sel)
	frame := screen.Frame{
		Title:    i18n.T(l.lang, "customs.title"),
		Items:    customsPanelItems(&l.settings, l.lang),
		Selected: *sel,
		Footer:   i18n.T(l.lang, "customs.footer"),
		Height:   screen.FrameHeight(l.height),
	}
	fmt.Fprint(l.out, screen.Render(frame, l.width))
	k, err := l.readKey()
	if err != nil {
		return "abort"
	}
	if k.Key == screen.KeyRune {
		return l.customsRuneKey(k.Rune, *sel)
	}
	switch k.Key {
	case screen.KeyUp:
		*sel = screen.MoveSteps(len(customs), *sel, -1)
	case screen.KeyDown:
		*sel = screen.MoveSteps(len(customs), *sel, 1)
	case screen.KeyEnter, screen.KeySpace:
		if !l.toggleCustomAt(*sel) {
			return "abort"
		}
	case screen.KeyQuit, screen.KeyEsc:
		return "menu"
	}
	return "customs"
}

// customsRuneKey handles single-letter panel actions; digits toggle rows.
func (l *screenLoop) customsRuneKey(r rune, sel int) string {
	switch r {
	case 'a', 'A':
		return l.customsAdd()
	case 'e', 'E':
		return l.customsEdit(sel)
	case 'd', 'D':
		return l.customsDelete(sel)
	case 'c', 'C':
		return l.pickEventView()
	}
	if idx := digitIndex(r, len(l.settings.CustomAlerts)); idx >= 0 {
		if !l.toggleCustomAt(idx) {
			return "abort"
		}
	}
	return "customs"
}

// toggleCustomAt flips the custom at position idx (no-op on empty rows),
// reporting false only when persisting fails.
func (l *screenLoop) toggleCustomAt(idx int) bool {
	if idx < 0 || idx >= len(l.settings.CustomAlerts) {
		return true
	}
	id := l.settings.CustomAlerts[idx].ID
	if !config.ToggleCustom(&l.settings, id) {
		return true
	}
	return l.persistSettings()
}

// customsDelete asks for confirmation inline, then removes and saves.
func (l *screenLoop) customsDelete(sel int) string {
	if sel < 0 || sel >= len(l.settings.CustomAlerts) {
		return "customs"
	}
	title := l.settings.CustomAlerts[sel].Title
	answer, ok := l.readLineInput(i18n.T(l.lang, "customs.title"), fmt.Sprintf(i18n.T(l.lang, "customs.confirmDelete"), title))
	if !ok {
		return "customs"
	}
	if answer != "y" && answer != "s" {
		return "customs"
	}
	id := l.settings.CustomAlerts[sel].ID
	if !config.RemoveCustom(&l.settings, id) {
		return "customs"
	}
	if !l.persistSettings() {
		return "abort"
	}
	return "customs"
}
