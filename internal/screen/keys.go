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
	// SGR mouse: ESC [ < Cb ; Cx ; Cy M/m. Wheel is Cb 64/65 + M,
	// left click is Cb 0 (+modifier bits) + M. Release (m) is
	// ignored so one press never activates twice.
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

// parseMouseSGR decodes one SGR mouse report from the head of buf.
// Incomplete reports return zero consumed so the caller waits for
// more bytes; malformed reports consume the introducer as unknown.
func parseMouseSGR(buf []byte) (ParsedKey, int) {
	end := -1
	for i := 3; i < len(buf); i++ {
		if buf[i] == 'M' || buf[i] == 'm' {
			end = i
			break
		}
		if (buf[i] < '0' || buf[i] > '9') && buf[i] != ';' {
			return ParsedKey{Key: KeyUnknown}, 3
		}
	}
	if end < 0 {
		return ParsedKey{Key: KeyUnknown}, 0
	}
	release := buf[end] == 'm'
	body := string(buf[3:end])
	cb, cx, cy := 0, 0, 0
	parts := 0
	num := 0
	hasNum := false
	for i := 0; i <= len(body); i++ {
		var c byte
		if i < len(body) {
			c = body[i]
		} else {
			c = ';'
		}
		if c >= '0' && c <= '9' {
			num = num*10 + int(c-'0')
			hasNum = true
			continue
		}
		if c != ';' {
			return ParsedKey{Key: KeyUnknown}, end + 1
		}
		if !hasNum {
			return ParsedKey{Key: KeyUnknown}, end + 1
		}
		switch parts {
		case 0:
			cb = num
		case 1:
			cx = num
		case 2:
			cy = num
		default:
			return ParsedKey{Key: KeyUnknown}, end + 1
		}
		parts++
		num = 0
		hasNum = false
	}
	if parts != 3 {
		return ParsedKey{Key: KeyUnknown}, end + 1
	}
	size := end + 1
	if release {
		return ParsedKey{Key: KeyUnknown}, size
	}
	// Wheel reports end with M and Cb 64 (up) or 65 (down).
	if cb == 64 {
		return ParsedKey{Key: KeyMouseWheel, MouseX: cx, MouseY: cy, Wheel: -1}, size
	}
	if cb == 65 {
		return ParsedKey{Key: KeyMouseWheel, MouseX: cx, MouseY: cy, Wheel: 1}, size
	}
	// Left-button press is Cb 0 plus optional modifier bits
	// (shift=4, alt=8, ctrl=16); low two bits 0 means no button drag.
	if cb&0x43 == 0 {
		return ParsedKey{Key: KeyMouseClick, MouseX: cx, MouseY: cy}, size
	}
	return ParsedKey{Key: KeyUnknown}, size
}
