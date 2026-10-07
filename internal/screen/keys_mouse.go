// SGR mouse reports: ESC [ < Cb ; Cx ; Cy M/m. Wheel is Cb 64/65 + M,
// motion is Cb with bit 5 + M, left click is Cb 0 (+modifier bits) + M.
// Release (m) is ignored so one press never activates twice.
package screen

import (
	"strconv"
	"strings"
)

// parseMouseSGR decodes one SGR mouse report from the head of buf.
// Incomplete reports return zero consumed so the caller waits for
// more bytes; malformed reports consume the introducer as unknown.
func parseMouseSGR(buf []byte) (ParsedKey, int) {
	end, malformed := mouseEnd(buf)
	if end < 0 {
		return ParsedKey{Key: KeyUnknown}, 0
	}
	if malformed {
		return ParsedKey{Key: KeyUnknown}, 3
	}
	cb, x, y, ok := mouseInts(string(buf[3:end]))
	if !ok {
		return ParsedKey{Key: KeyUnknown}, end + 1
	}
	size := end + 1
	if buf[end] == 'm' {
		return ParsedKey{Key: KeyUnknown}, size
	}
	if cb == 64 {
		return ParsedKey{Key: KeyMouseWheel, MouseX: x, MouseY: y, Wheel: -1}, size
	}
	if cb == 65 {
		return ParsedKey{Key: KeyMouseWheel, MouseX: x, MouseY: y, Wheel: 1}, size
	}
	// Motion (hover or drag) reports set bit 5 (32) with M. Hover only
	// moves the selection; views never activate from it.
	if cb&32 != 0 {
		return ParsedKey{Key: KeyMouseMotion, MouseX: x, MouseY: y}, size
	}
	// Left-button press is Cb 0 plus optional modifier bits
	// (shift=4, alt=8, ctrl=16); low two bits 0 means no button drag.
	if cb&0x43 == 0 {
		return ParsedKey{Key: KeyMouseClick, MouseX: x, MouseY: y}, size
	}
	return ParsedKey{Key: KeyUnknown}, size
}

// mouseEnd scans the SGR body for its M/m terminator, returning its
// index. -1 means more bytes may still complete it; malformed is set
// when a byte outside digits and ';' can never belong to a report.
func mouseEnd(buf []byte) (int, bool) {
	for i := 3; i < len(buf); i++ {
		if buf[i] == 'M' || buf[i] == 'm' {
			return i, false
		}
		if !isMouseBodyByte(buf[i]) {
			return i, true
		}
	}
	return -1, false
}

func isMouseBodyByte(c byte) bool {
	return (c >= '0' && c <= '9') || c == ';'
}

// mouseInts parses a "Cb;Cx;Cy" body into button and coordinates.
func mouseInts(body string) (int, int, int, bool) {
	parts := strings.Split(body, ";")
	if len(parts) != 3 {
		return 0, 0, 0, false
	}
	nums := make([]int, 0, 3)
	for _, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return 0, 0, 0, false
		}
		nums = append(nums, n)
	}
	return nums[0], nums[1], nums[2], true
}
