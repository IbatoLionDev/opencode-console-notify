// Fullscreen customs editor: event picker plus single-line text input
// with echo and backspace, staying inside the alternate screen.
package tui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/IbatoLionDev/opencode-console-notify/internal/config"
	"github.com/IbatoLionDev/opencode-console-notify/internal/i18n"
	"github.com/IbatoLionDev/opencode-console-notify/internal/screen"
)

// customsTitleKey is the frame title for every custom prompt so the key
// stays in one place instead of repeated across add/edit flows.
const customsTitleKey = "customs.title"

// readByte returns one input byte, draining buffered key bytes first so
// text input never loses bytes to a previous split escape sequence.
func (l *screenLoop) readByte() (byte, bool) {
	if len(l.pending) > 0 {
		b := l.pending[0]
		l.pending = l.pending[1:]
		return b, true
	}
	var one [1]byte
	n, err := l.in.Read(one[:])
	if n == 0 || err != nil {
		return 0, false
	}
	return one[0], true
}

// readRune decodes one rune incrementally so backspace never splits a
// multibyte sequence.
func (l *screenLoop) readRune() (rune, bool) {
	var buf [4]byte
	n := 0
	for n < len(buf) {
		b, ok := l.readByte()
		if !ok {
			return 0, false
		}
		buf[n] = b
		n++
		if r, size := utf8.DecodeRune(buf[:n]); size == n {
			return r, true
		}
	}
	return utf8.RuneError, true
}

// promptFrame renders the text prompt as a frame so input never leaves
// the alternate screen; every keystroke redraws with the buffer.
func (l *screenLoop) promptFrame(title, prompt string, buf []rune) {
	frame := screen.Frame{
		Title:    title,
		Items:    []screen.Item{{Label: prompt + string(buf)}},
		Selected: 0,
		Footer:   i18n.T(l.lang, "customs.inputFooter"),
		Height:   screen.FrameHeight(l.height),
	}
	fmt.Fprint(l.out, screen.Render(frame, l.width))
	// Render ends with footer + trailing newline, so reposition the
	// hardware cursor inside the input row after prompt+typed buffer.
	col := screen.InputCursorCol(prompt, len(buf), l.width)
	fmt.Fprint(l.out, screen.MoveCursor(screen.InputRow, col))
	// Drawn white block marks the insertion point; the hardware cursor
	// stays hidden (EnterFrame hides, LeaveFrame restores).
	fmt.Fprint(l.out, screen.CursorBlock())
}

// readLineInput reads one line inside a prompt frame with echo and
// backspace. Enter submits (trimmed), Esc aborts (ok=false). The drawn
// white block marks the insertion point while typing; the hardware
// cursor stays hidden (HideCursor defer is safety — EnterFrame already
// hides, LeaveFrame restores).
func (l *screenLoop) readLineInput(frameTitle, prompt string) (string, bool) {
	defer fmt.Fprint(l.out, screen.HideCursor)
	var runes []rune
	for {
		l.promptFrame(frameTitle, prompt, runes)
		r, ok := l.readRune()
		if !ok {
			return "", false
		}
		switch r {
		case '\r', '\n':
			return strings.TrimSpace(string(runes)), true
		case 0x1b:
			return "", false
		case 0x7f, '\b':
			if len(runes) > 0 {
				runes = runes[:len(runes)-1]
			}
		default:
			if r >= 0x20 {
				runes = append(runes, r)
			}
		}
	}
}

// pickEvent renders the catalog as a scrollable frame and returns the
// chosen index, or -1 on cancel. Noisy events carry a "!" state mark.
func (l *screenLoop) pickEvent(preselect string) int {
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
		if k.Key == screen.KeyMouseWheel {
			sel = screen.MoveSteps(len(items), sel, k.Wheel)
			continue
		}
		if k.Key == screen.KeyMouseClick {
			idx, ok := screen.ItemIndexAtRow(items, sel, screen.FrameHeight(l.height), k.MouseY)
			if !ok {
				continue
			}
			return idx
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

// customsAdd runs create: event picker, then title and body prompts.
func (l *screenLoop) customsAdd() string {
	events := config.KnownEvents()
	idx := l.pickEvent("")
	if idx < 0 {
		return "customs"
	}
	title, ok := l.readLineInput(i18n.T(l.lang, customsTitleKey), i18n.T(l.lang, "customs.titleLabel")+": ")
	if !ok || strings.TrimSpace(title) == "" {
		return "customs"
	}
	body, ok := l.readLineInput(i18n.T(l.lang, customsTitleKey), i18n.T(l.lang, "customs.bodyLabel")+" ("+i18n.T(l.lang, "customs.optional")+"): ")
	if !ok {
		return "customs"
	}
	if _, err := config.AddCustom(&l.settings, events[idx].Type, title, body); err != nil {
		fmt.Fprintf(l.out, errorFormat, err)
		return "abort"
	}
	if !l.persistSettings() {
		return "abort"
	}
	return "customs"
}

// customsEdit re-prompts title and body (empty keeps current) plus the
// event picker preselected on the current event.
func (l *screenLoop) customsEdit(sel int) string {
	if sel < 0 || sel >= len(l.settings.CustomAlerts) {
		return "customs"
	}
	current := l.settings.CustomAlerts[sel]
	title, ok := l.readLineInput(i18n.T(l.lang, customsTitleKey), i18n.T(l.lang, "customs.titleLabel")+" ["+current.Title+"]: ")
	if !ok {
		return "customs"
	}
	if strings.TrimSpace(title) == "" {
		title = current.Title
	}
	body, ok := l.readLineInput(i18n.T(l.lang, customsTitleKey), i18n.T(l.lang, "customs.bodyLabel")+" ["+current.Body+"]: ")
	if !ok {
		return "customs"
	}
	if strings.TrimSpace(body) == "" && current.Body != "" {
		body = current.Body
	}
	idx := l.pickEvent(current.Event)
	if idx < 0 {
		return "customs"
	}
	events := config.KnownEvents()
	kept, err := config.UpdateCustom(&l.settings, current.ID, events[idx].Type, title, body)
	if err != nil {
		fmt.Fprintf(l.out, errorFormat, err)
		return "abort"
	}
	if !kept {
		return "customs"
	}
	if !l.persistSettings() {
		return "abort"
	}
	return "customs"
}
