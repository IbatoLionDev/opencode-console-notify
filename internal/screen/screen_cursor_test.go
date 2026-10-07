package screen

import "testing"

func TestMoveCursorFormatAndClamp(t *testing.T) {
	cases := []struct {
		name string
		row  int
		col  int
		want string
	}{
		{name: "input row after prompt", row: 3, col: 10, want: "\x1b[3;10H"},
		{name: "origin", row: 1, col: 1, want: "\x1b[1;1H"},
		{name: "row clamps to 1", row: 0, col: 5, want: "\x1b[1;5H"},
		{name: "col clamps to 1", row: 3, col: -2, want: "\x1b[3;1H"},
		{name: "both clamp to 1", row: -4, col: 0, want: "\x1b[1;1H"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := MoveCursor(tt.row, tt.col); got != tt.want {
				t.Fatalf("MoveCursor(%d, %d) = %q, want %q", tt.row, tt.col, got, tt.want)
			}
		})
	}
	if InputRow != 3 {
		t.Fatalf("InputRow = %d, want 3 (title=1, border=2, input=3)", InputRow)
	}
}

func TestCursorBlockIsWhiteCell(t *testing.T) {
	if got := CursorBlock(); got != BgWhite+" "+Reset {
		t.Fatalf("CursorBlock() = %q, want BgWhite+space+Reset", got)
	}
}

func TestInputCursorCol(t *testing.T) {
	const titlePrompt = "Title: "
	cases := []struct {
		name   string
		prompt string
		typed  int
		width  int
		want   int
	}{
		{name: "empty buffer", prompt: titlePrompt, typed: 0, width: 40, want: 1 + 2 + 7},
		{name: "with typed", prompt: titlePrompt, typed: 5, width: 40, want: 1 + 2 + 7 + 5},
		{name: "clamps at width", prompt: titlePrompt, typed: 100, width: 40, want: 40},
		{name: "narrow width clamps to 10", prompt: "T: ", typed: 100, width: 4, want: 10},
		{name: "multibyte prompt counts runes", prompt: "Título: ", typed: 0, width: 40, want: 1 + 2 + 8},
		{name: "negative typed clamps low", prompt: "", typed: -10, width: 40, want: 1},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := InputCursorCol(tt.prompt, tt.typed, tt.width); got != tt.want {
				t.Fatalf("InputCursorCol(%q, %d, %d) = %d, want %d", tt.prompt, tt.typed, tt.width, got, tt.want)
			}
		})
	}
}
