// Fullscreen customs event picker: scrollable catalog of subscribable
// events with type-to-filter search plus click and wheel support. Text
// input lives in customs_input.go; add/edit flows in customs_editor.go.
package tui

import (
	"fmt"
	"strings"

	"github.com/IbatoLionDev/opencode-console-notify/internal/config"
	"github.com/IbatoLionDev/opencode-console-notify/internal/i18n"
	"github.com/IbatoLionDev/opencode-console-notify/internal/screen"
)

// pickerRows builds the visible catalog rows for one filter query,
// returning the rows plus the event index behind each row. Matching is
// a strict contiguous substring (case-insensitive) over the event type
// or description, from any position; an empty query keeps everything.
func (l *screenLoop) pickerRows(query string) ([]screen.Item, []int) {
	events := config.KnownEvents()
	q := strings.ToLower(query)
	items := make([]screen.Item, 0, len(events))
	indexMap := make([]int, 0, len(events))
	for i, e := range events {
		state := ""
		if e.Noisy {
			state = "!"
		}
		desc := e.DescEn
		if l.lang == "es" {
			desc = e.DescEs
		}
		if q != "" && !strings.Contains(strings.ToLower(e.Type), q) && !strings.Contains(strings.ToLower(desc), q) {
			continue
		}
		items = append(items, screen.Item{Label: e.Type, Detail: desc, State: state})
		indexMap = append(indexMap, i)
	}
	return items, indexMap
}

// pickerTypeKey updates the filter query from one keypress: printable
// runes append (q included, which types here instead of quitting),
// Backspace trims rune-aware. It reports the new query and whether the
// key typed.
func pickerTypeKey(k screen.ParsedKey, query string) (string, bool) {
	if k.Key != screen.KeyRune && k.Key != screen.KeyQuit {
		return query, false
	}
	if k.Rune == 0x7f || k.Rune == '\b' {
		if query == "" {
			return query, false
		}
		runes := []rune(query)
		return string(runes[:len(runes)-1]), true
	}
	if k.Rune < 0x20 {
		return query, false
	}
	return query + string(k.Rune), true
}

// pickerMouse handles wheel (scroll) and click (pick) for the catalog.
// It returns the new position, the picked event index (-1 while
// browsing), and whether the key was a mouse report.
func pickerMouse(k screen.ParsedKey, sel int, items []screen.Item, indexMap []int, height int) (int, int, bool) {
	if k.Key == screen.KeyMouseWheel {
		return screen.MoveSteps(len(items), sel, k.Wheel), -1, true
	}
	if k.Key == screen.KeyMouseMotion {
		if idx, ok := screen.ItemIndexAtRow(items, sel, height, k.MouseY); ok {
			return idx, -1, true
		}
		return sel, -1, true
	}
	if k.Key != screen.KeyMouseClick {
		return sel, -1, false
	}
	pos, ok := screen.ItemIndexAtRow(items, sel, height, k.MouseY)
	if !ok {
		return sel, -1, true
	}
	return pos, indexMap[pos], true
}

// pickerNavKey moves the selection or picks the highlighted row for one
// navigation key. Enter/Space on an empty list stays put.
func pickerNavKey(k screen.ParsedKey, sel int, items []screen.Item, indexMap []int) (int, int, bool) {
	switch k.Key {
	case screen.KeyUp:
		return screen.MoveSteps(len(items), sel, -1), -1, true
	case screen.KeyDown:
		return screen.MoveSteps(len(items), sel, 1), -1, true
	case screen.KeyEnter, screen.KeySpace:
		if len(items) == 0 {
			return sel, -1, true
		}
		return sel, indexMap[sel], true
	}
	return sel, -1, false
}

// pickerFooter renders the picker hint plus the live filter status:
// Filter "sess" 3/48 while typing, the plain hint otherwise.
func pickerFooter(lang, query string, shown, total int) string {
	base := i18n.T(lang, "customs.pickFooter")
	if query == "" {
		return base
	}
	return fmt.Sprintf("%s | Filter %q %d/%d", base, query, shown, total)
}

// pickerStep routes one picker keypress through mouse, typing, Esc and
// navigation, mutating the query and selection in place. It returns the
// picked event index and true when the picker is done (picked, or backed
// out with -1).
func (l *screenLoop) pickerStep(k screen.ParsedKey, query *string, sel *int, items []screen.Item, indexMap []int) (int, bool) {
	if ns, pk, ok := pickerMouse(k, *sel, items, indexMap, screen.FrameHeight(l.height)); ok {
		*sel = ns
		if pk >= 0 {
			return pk, true
		}
		return -1, false
	}
	if nq, ok := pickerTypeKey(k, *query); ok {
		if nq != *query {
			*query = nq
			*sel = 0
		}
		return -1, false
	}
	if k.Key == screen.KeyEsc {
		if *query != "" {
			*query = ""
			*sel = 0
			return -1, false
		}
		return -1, true
	}
	if ns, pk, ok := pickerNavKey(k, *sel, items, indexMap); ok {
		*sel = ns
		if pk >= 0 {
			return pk, true
		}
	}
	return -1, false
}

// pickEvent renders the catalog as a scrollable filterable frame and
// returns the chosen event index, or -1 on cancel. Noisy events carry
// a "!" state mark.
func (l *screenLoop) pickEvent(preselect string) int {
	sel := 0
	events := config.KnownEvents()
	for i, e := range events {
		if e.Type == preselect {
			sel = i
			break
		}
	}
	total := len(events)
	query := ""
	for {
		items, indexMap := l.pickerRows(query)
		sel = screen.ClampSelection(len(items), sel)
		frame := screen.Frame{
			Title:    i18n.T(l.lang, "customs.pickEvent"),
			Items:    items,
			Selected: sel,
			Footer:   pickerFooter(l.lang, query, len(items), total),
			Height:   screen.FrameHeight(l.height),
		}
		fmt.Fprint(l.out, screen.Render(frame, l.width))
		k, err := l.readKey()
		if err != nil {
			return -1
		}
		if picked, done := l.pickerStep(k, &query, &sel, items, indexMap); done {
			return picked
		}
	}
}

// pickEventView opens the read-only catalog from the panel (the fifth
// option): any key walks back, Enter included.
func (l *screenLoop) pickEventView() string {
	l.pickEvent("")
	return "customs"
}
