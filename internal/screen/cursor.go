// Drawn cursor helpers: the hardware cursor stays hidden during text
// input, so the prompt frame draws a white block at the insertion point
// and parks the hardware cursor inside the input row after every render.
package screen

import "fmt"

// InputRow is the 1-indexed terminal row of the input row in a
// single-item prompt frame (title=1, top border=2, input=3).
const InputRow = 3

// CursorBlock returns the drawn insertion-point block: a white-bg ASCII
// space (solid white cell, zero wide-rune risk) followed by Reset. The
// hardware cursor stays hidden during text input; this block is written
// at the insertion point after every prompt-frame render instead.
func CursorBlock() string {
	return BgWhite + " " + Reset
}

// MoveCursor returns the ANSI CUP sequence placing the hardware cursor
// at the given 1-indexed row and column, clamping each to >= 1.
func MoveCursor(row, col int) string {
	if row < 1 {
		row = 1
	}
	if col < 1 {
		col = 1
	}
	return fmt.Sprintf("\x1b[%d;%dH", row, col)
}

// InputCursorCol returns the 1-indexed column just after prompt+typed in
// the input row: 1 + marker width ("> ") + prompt runes + typed runes,
// clamped to [1, effectiveWidth] where effectiveWidth mirrors Render.
func InputCursorCol(prompt string, typed int, width int) int {
	if width < 10 {
		width = 10
	}
	col := 1 + 2 + len([]rune(prompt)) + typed
	if col < 1 {
		return 1
	}
	if col > width {
		return width
	}
	return col
}
