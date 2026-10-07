// Fullscreen customs text input: single-line echo with backspace,
// staying inside the alternate screen. Picker and add/edit flows live
// in customs_picker.go and customs_editor.go.
package tui

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/IbatoLionDev/opencode-console-notify/internal/i18n"
	"github.com/IbatoLionDev/opencode-console-notify/internal/screen"
)

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

// swallowMouseReport discards one SGR mouse report whose ESC was already
// consumed. It reports whether pending held a complete "[<...M/m"
// report; anything else (a lone ESC press) is left for the caller.
func (l *screenLoop) swallowMouseReport() bool {
	for {
		if len(l.pending) < 2 || l.pending[0] != '[' || l.pending[1] != '<' {
			return false
		}
		for i := 2; i < len(l.pending); i++ {
			c := l.pending[i]
			if c == 'M' || c == 'm' {
				l.pending = l.pending[i+1:]
				return true
			}
			if (c < '0' || c > '9') && c != ';' {
				return false
			}
		}
		var chunk [16]byte
		n, err := l.in.Read(chunk[:])
		if n > 0 {
			l.pending = append(l.pending, chunk[:n]...)
			continue
		}
		if err != nil {
			return false
		}
	}
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
			// A mouse report starts with ESC too: swallow it so hover
			// and clicks neither abort the prompt nor dirty the buffer.
			if l.swallowMouseReport() {
				continue
			}
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
