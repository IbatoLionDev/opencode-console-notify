// Fullscreen language view: English/Spanish picker rendered as a frame.
package tui

import (
	"fmt"

	"github.com/IbatoLionDev/opencode-console-notify/internal/i18n"
	"github.com/IbatoLionDev/opencode-console-notify/internal/screen"
)

func (l *screenLoop) runLanguage(sel *int) string {
	items := []screen.Item{
		{Label: "English", State: currentMark(l.lang == "en", l.lang)},
		{Label: "Español", State: currentMark(l.lang == "es", l.lang)},
	}
	if *sel < 0 || *sel >= len(items) {
		*sel = langIndex(l.lang)
	}
	frame := screen.Frame{
		Title:    i18n.T(l.lang, "lang.title"),
		Items:    items,
		Selected: *sel,
		Footer:   i18n.T(l.lang, "footer.back"),
		Height:   screen.FrameHeight(l.height),
	}
	fmt.Fprint(l.out, screen.Render(frame, l.width))
	k, err := l.readKey()
	if err != nil {
		return "abort"
	}
	if next, ok := l.languageMouse(k, items, sel); ok {
		return next
	}
	if k.Key == screen.KeyRune {
		if k.Rune == '1' || k.Rune == 'e' || k.Rune == 'E' {
			return l.setLanguage("en")
		}
		if k.Rune == '2' {
			return l.setLanguage("es")
		}
	}
	switch k.Key {
	case screen.KeyEnter, screen.KeySpace:
		if langIndex(l.lang) == 0 {
			return l.setLanguage("es")
		}
		return l.setLanguage("en")
	case screen.KeyQuit, screen.KeyEsc:
		return "menu"
	}
	return "language"
}

// languageMouse handles one mouse report for the language view,
// reporting the next view and whether the key was consumed. Wheel
// flips between languages, hover moves the highlight, click picks
// the row under the cursor.
func (l *screenLoop) languageMouse(k screen.ParsedKey, items []screen.Item, sel *int) (string, bool) {
	if k.Key == screen.KeyMouseWheel {
		if k.Wheel < 0 {
			return l.setLanguage("en"), true
		}
		return l.setLanguage("es"), true
	}
	if k.Key == screen.KeyMouseMotion {
		if idx, ok := screen.ItemIndexAtRow(items, *sel, screen.FrameHeight(l.height), k.MouseY); ok {
			*sel = idx
		}
		return "language", true
	}
	if k.Key != screen.KeyMouseClick {
		return "", false
	}
	idx, ok := screen.ItemIndexAtRow(items, *sel, screen.FrameHeight(l.height), k.MouseY)
	if !ok {
		return "language", true
	}
	*sel = idx
	if idx == 0 {
		return l.setLanguage("en"), true
	}
	return l.setLanguage("es"), true
}

func langIndex(lang string) int {
	if lang == "es" {
		return 1
	}
	return 0
}

func currentMark(active bool, lang string) string {
	if active {
		return i18n.T(lang, "lang.current")
	}
	return ""
}

func (l *screenLoop) setLanguage(lang string) string {
	l.lang = lang
	l.settings.Lang = lang
	if !l.persistSettings() {
		return "abort"
	}
	return "language"
}
