// Fullscreen customs event picker: scrollable catalog of subscribable
// events with click and wheel support. Text input lives in
// customs_input.go; add/edit flows in customs_editor.go.
package tui

import (
	"fmt"

	"github.com/IbatoLionDev/opencode-console-notify/internal/config"
	"github.com/IbatoLionDev/opencode-console-notify/internal/i18n"
	"github.com/IbatoLionDev/opencode-console-notify/internal/screen"
)

// pickerItems builds the catalog rows, preselecting the current event.
func (l *screenLoop) pickerItems(preselect string) ([]screen.Item, int) {
	events := config.KnownEvents()
	items := make([]screen.Item, 0, len(events))
	sel := 0
	for i, e := range events {
		if e.Type == preselect {
			sel = i
		}
		state := ""
		if e.Noisy {
			state = "!"
		}
		desc := e.DescEn
		if l.lang == "es" {
			desc = e.DescEs
		}
		items = append(items, screen.Item{Label: e.Type, Detail: desc, State: state})
	}
	return items, sel
}

// pickerMouse handles wheel (scroll) and click (pick) for the catalog.
// It returns the new selection, the picked index (-1 while scrolling),
// and whether the key was a mouse report.
func pickerMouse(k screen.ParsedKey, sel int, items []screen.Item, height int) (int, int, bool) {
	if k.Key == screen.KeyMouseWheel {
		return screen.MoveSteps(len(items), sel, k.Wheel), -1, true
	}
	if k.Key != screen.KeyMouseClick {
		return sel, -1, false
	}
	idx, ok := screen.ItemIndexAtRow(items, sel, height, k.MouseY)
	if !ok {
		return sel, -1, true
	}
	return idx, idx, true
}

// pickEvent renders the catalog as a scrollable frame and returns the
// chosen index, or -1 on cancel. Noisy events carry a "!" state mark.
func (l *screenLoop) pickEvent(preselect string) int {
	items, sel := l.pickerItems(preselect)
	for {
		frame := screen.Frame{
			Title:    i18n.T(l.lang, "customs.pickEvent"),
			Items:    items,
			Selected: sel,
			Footer:   i18n.T(l.lang, "footer.back"),
			Height:   screen.FrameHeight(l.height),
		}
		fmt.Fprint(l.out, screen.Render(frame, l.width))
		k, err := l.readKey()
		if err != nil {
			return -1
		}
		if ns, picked, ok := pickerMouse(k, sel, items, screen.FrameHeight(l.height)); ok {
			if picked >= 0 {
				return picked
			}
			sel = ns
			continue
		}
		if k.Key == screen.KeyRune {
			return -1
		}
		switch k.Key {
		case screen.KeyUp:
			sel = screen.MoveSteps(len(items), sel, -1)
		case screen.KeyDown:
			sel = screen.MoveSteps(len(items), sel, 1)
		case screen.KeyEnter, screen.KeySpace:
			return sel
		case screen.KeyQuit, screen.KeyEsc:
			return -1
		}
	}
}

// pickEventView opens the read-only catalog from the panel (the fifth
// option): any key walks back, Enter included.
func (l *screenLoop) pickEventView() string {
	l.pickEvent("")
	return "customs"
}
