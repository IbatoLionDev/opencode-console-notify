// Fullscreen alerts view: the four default toggles rendered as a frame.
package tui

import (
	"fmt"

	"github.com/IbatoLionDev/opencode-console-notify/internal/config"
	"github.com/IbatoLionDev/opencode-console-notify/internal/i18n"
	"github.com/IbatoLionDev/opencode-console-notify/internal/screen"
)

// alertScreenItems maps the shared AlertItems domain over screen rows so
// the four alerts are defined once (tui_renders.go owns the mapping).
func alertScreenItems(s *config.Settings, lang string) []screen.Item {
	fields := AlertItems(s, lang)
	items := make([]screen.Item, 0, len(fields))
	for _, f := range fields {
		items = append(items, screen.Item{
			Label:  f.Name,
			Detail: f.Trigger,
			State:  i18n.T(lang, boolStateKey(f.Enabled)),
		})
	}
	return items
}

func boolStateKey(on bool) string {
	if on {
		return "state.on"
	}
	return "state.off"
}

func alertSetting(s *config.Settings, idx int) *bool {
	switch idx {
	case 0:
		return &s.Alerts.SessionIdle
	case 1:
		return &s.Alerts.SessionError
	case 2:
		return &s.Alerts.PermissionAsked
	}
	return &s.Alerts.QuestionAsked
}

func (l *screenLoop) toggleAlert(idx int) bool {
	*alertSetting(&l.settings, idx) = !*alertSetting(&l.settings, idx)
	return l.persistSettings()
}

func (l *screenLoop) runAlerts(sel *int) string {
	items := alertScreenItems(&l.settings, l.lang)
	*sel = screen.ClampSelection(len(items), *sel)
	frame := screen.Frame{
		Title:    i18n.T(l.lang, "alerts.title"),
		Items:    items,
		Selected: *sel,
		Footer:   i18n.T(l.lang, "footer.toggle"),
		Height:   screen.FrameHeight(l.height),
	}
	fmt.Fprint(l.out, screen.Render(frame, l.width))
	k, err := l.readKey()
	if err != nil {
		return "abort"
	}
	if k.Key == screen.KeyRune {
		if idx := digitIndex(k.Rune, len(items)); idx >= 0 {
			if !l.toggleAlert(idx) {
				return "abort"
			}
		}
		return "alerts"
	}
	switch k.Key {
	case screen.KeyUp:
		*sel = screen.MoveSteps(len(items), *sel, -1)
	case screen.KeyDown:
		*sel = screen.MoveSteps(len(items), *sel, 1)
	case screen.KeyEnter, screen.KeySpace:
		if !l.toggleAlert(*sel) {
			return "abort"
		}
	case screen.KeyQuit, screen.KeyEsc:
		return "menu"
	}
	return "alerts"
}

// digitIndex maps 1-based digit runes to item indexes, -1 when outside.
func digitIndex(r rune, n int) int {
	if r < '1' || int(r-'1') >= n {
		return -1
	}
	return int(r - '1')
}
