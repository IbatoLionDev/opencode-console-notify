// Key model for the fullscreen loop: arrows and j/k navigate, Enter
// selects, Space toggles, digits and letters double as shortcuts.
package screen

// Key is one parsed keypress.
type Key int

const (
	KeyUnknown Key = iota
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyEnter
	KeySpace
	KeyEsc
	KeyQuit
	KeyRune
	KeyMouseClick
	KeyMouseWheel
)

// ParsedKey is a Key plus the rune for KeyRune (shortcuts like 1-5, l, q).
// Mouse clicks carry 1-indexed terminal coords in MouseX/MouseY;
// wheel events carry Wheel (-1 up, +1 down) plus the coords.
type ParsedKey struct {
	Key    Key
	Rune   rune
	MouseX int
	MouseY int
	Wheel  int
}

// ParseKey decodes one keypress from the head of buf and reports how many
// bytes it consumed. A lone ESC byte means Esc only when nothing follows
// it in this chunk; split escape sequences across reads stay KeyUnknown
// with zero consumed so the caller waits for more bytes.
func ParseKey(buf []byte) (ParsedKey, int) {
	if len(buf) == 0 {
		return ParsedKey{Key: KeyUnknown}, 0
	}
	switch buf[0] {
	case '\r', '\n':
		return ParsedKey{Key: KeyEnter}, 1
	case ' ':
		return ParsedKey{Key: KeySpace}, 1
	case '\x1b':
		return parseEscape(buf)
	}
	r, size := decodeRune(buf)
	switch r {
	case 'q', 'Q':
		return ParsedKey{Key: KeyQuit, Rune: r}, size
	case 'k', 'K':
		return ParsedKey{Key: KeyUp, Rune: r}, size
	case 'j', 'J':
		return ParsedKey{Key: KeyDown, Rune: r}, size
	case 'h', 'H':
		return ParsedKey{Key: KeyLeft, Rune: r}, size
	case 'l', 'L':
		return ParsedKey{Key: KeyRight, Rune: r}, size
	}
	return ParsedKey{Key: KeyRune, Rune: r}, size
}

func parseEscape(buf []byte) (ParsedKey, int) {
	if len(buf) == 1 {
		return ParsedKey{Key: KeyEsc}, 1
	}
	if len(buf) == 2 {
		return ParsedKey{Key: KeyUnknown}, 0
	}
	// SGR mouse: ESC [ < Cb ; Cx ; Cy M/m (see keys_mouse.go).
	if buf[1] == '[' && len(buf) > 3 && buf[2] == '<' {
		return parseMouseSGR(buf)
	}
	if buf[1] != '[' && buf[1] != 'O' {
		return ParsedKey{Key: KeyEsc}, 1
	}
	switch buf[2] {
	case 'A':
		return ParsedKey{Key: KeyUp}, 3
	case 'B':
		return ParsedKey{Key: KeyDown}, 3
	case 'C':
		return ParsedKey{Key: KeyRight}, 3
	case 'D':
		return ParsedKey{Key: KeyLeft}, 3
	}
	// Unknown CSI sequence: swallow the introducer so the loop never spins.
	return ParsedKey{Key: KeyUnknown}, 3
}

// decodeRune decodes one UTF-8 rune without importing unicode/utf8.
// Invalid bytes become RuneError of size 1 so the loop always advances.
func decodeRune(buf []byte) (rune, int) {
	c := buf[0]
	if c < 0x80 {
		return rune(c), 1
	}
	var size int
	var min rune
	switch {
	case c>>5 == 0x06:
		size, min = 2, 0x80
	case c>>4 == 0x0E:
		size, min = 3, 0x800
	case c>>3 == 0x1E:
		size, min = 4, 0x10000
	default:
		return 0xFFFD, 1
	}
	if len(buf) < size {
		return 0xFFFD, 1
	}
	r := rune(c & byte(0xFF>>(size+1)))
	for i := 1; i < size; i++ {
		if buf[i]>>6 != 0x02 {
			return 0xFFFD, 1
		}
		r = r<<6 | rune(buf[i]&0x3F)
	}
	if r < min || (r >= 0xD800 && r <= 0xDFFF) {
		return 0xFFFD, 1
	}
	return r, size
}
