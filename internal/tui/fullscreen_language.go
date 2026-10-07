// Fullscreen language view: English/Spanish picker rendered as a frame.
package tui

import (
	"fmt"

	"github.com/IbatoLionDev/opencode-console-notify/internal/i18n"
	"github.com/IbatoLionDev/opencode-console-notify/internal/screen"
)

func (l *screenLoop) runLanguage() string {
	items := []screen.Item{
		{Label: "English", State: currentMark(l.lang == "en", l.lang)},
		{Label: "Español", State: currentMark(l.lang == "es", l.lang)},
	}
	frame := screen.Frame{
		Title:    i18n.T(l.lang, "lang.title"),
		Items:    items,
		Selected: screen.ClampSelection(len(items), langIndex(l.lang)),
		Footer:   i18n.T(l.lang, "footer.back"),
		Height:   screen.FrameHeight(l.height),
	}
	fmt.Fprint(l.out, screen.Render(frame, l.width))
	k, err := l.readKey()
	if err != nil {
		return "abort"
	}
	if k.Key == screen.KeyMouseWheel {
		if k.Wheel < 0 {
			return l.setLanguage("en")
		}
		return l.setLanguage("es")
	}
	if k.Key == screen.KeyMouseClick {
		idx, ok := screen.ItemIndexAtRow(items, screen.ClampSelection(len(items), langIndex(l.lang)), screen.FrameHeight(l.height), k.MouseY)
		if !ok {
			return "language"
		}
		if idx == 0 {
			return l.setLanguage("en")
		}
		return l.setLanguage("es")
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
