package screen

import (
	"strings"
	"testing"
)

func TestViewportKeepsSelectionVisible(t *testing.T) {
	items := []Item{{Label: "a"}, {Label: "b"}, {Label: "c"}, {Label: "d"}, {Label: "e"}}
	visible, offset := Viewport(items, 4, 2)
	if offset == 0 {
		t.Fatal("last item with height 2 must scroll (offset > 0)")
	}
	if visible[len(visible)-1].Label != "e" {
		t.Fatalf("selection must stay visible, got %v", visible)
	}
	visible, offset = Viewport(items, 0, 2)
	if offset != 0 || visible[0].Label != "a" {
		t.Fatalf("top selection must not scroll, got offset=%d %v", offset, visible)
	}
	if _, offset := Viewport(items, 2, 0); offset != 0 {
		t.Fatal("height 0 disables scrolling")
	}
}

func TestViewportCountsDetailLines(t *testing.T) {
	items := []Item{{Label: "a", Detail: "x"}, {Label: "b", Detail: "y"}, {Label: "c"}}
	visible, _ := Viewport(items, 2, 3)
	found := false
	for _, it := range visible {
		if it.Label == "c" {
			found = true
		}
	}
	if !found {
		t.Fatalf("selection c must fit in height 3 with details counted, got %v", visible)
	}
}

func TestParseKeyArrowsAndControls(t *testing.T) {
	cases := []struct {
		name string
		buf  []byte
		want Key
		size int
	}{
		{name: "up", buf: []byte{0x1b, '[', 'A'}, want: KeyUp, size: 3},
		{name: "down", buf: []byte{0x1b, '[', 'B'}, want: KeyDown, size: 3},
		{name: "enter cr", buf: []byte{'\r'}, want: KeyEnter, size: 1},
		{name: "enter lf", buf: []byte{'\n'}, want: KeyEnter, size: 1},
		{name: "space", buf: []byte{' '}, want: KeySpace, size: 1},
		{name: "esc alone", buf: []byte{0x1b}, want: KeyEsc, size: 1},
		{name: "j down", buf: []byte{'j'}, want: KeyDown, size: 1},
		{name: "k up", buf: []byte{'k'}, want: KeyUp, size: 1},
		{name: "q quit", buf: []byte{'q'}, want: KeyQuit, size: 1},
		{name: "split escape waits", buf: []byte{0x1b, '['}, want: KeyUnknown, size: 0},
		{name: "empty", buf: nil, want: KeyUnknown, size: 0},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got, size := ParseKey(tt.buf)
			if got.Key != tt.want || size != tt.size {
				t.Fatalf("got (%v, %d), want (%v, %d)", got.Key, size, tt.want, tt.size)
			}
		})
	}
}

func TestParseKeyRunes(t *testing.T) {
	got, size := ParseKey([]byte{'5'})
	if got.Key != KeyRune || got.Rune != '5' || size != 1 {
		t.Fatalf("digit must be a rune shortcut, got %+v size %d", got, size)
	}
	if got, _ := ParseKey([]byte("ñ")); got.Rune != 'ñ' {
		t.Fatalf("multibyte rune must decode, got %+v", got)
	}
}

func TestRenderMarksSelection(t *testing.T) {
	f := Frame{
		Title:    "Config",
		Items:    []Item{{Label: "one"}, {Label: "two", Detail: "trigger"}},
		Selected: 1,
		Footer:   "keys",
		Height:   10,
	}
	out := Render(f, 40)
	if !strings.Contains(out, "> ") {
		t.Fatal("render must mark the selected row")
	}
	if !strings.Contains(out, "trigger") {
		t.Fatal("render must include detail lines")
	}
	if !strings.Contains(out, "keys") {
		t.Fatal("render must include the footer")
	}
	if strings.Count(out, "> ") != 1 {
		t.Fatal("exactly one row must be marked selected")
	}
}

func TestMoveStepsWraps(t *testing.T) {
	if got := MoveSteps(5, 0, -1); got != 4 {
		t.Fatalf("up from 0 must wrap to 4, got %d", got)
	}
	if got := MoveSteps(5, 4, 1); got != 0 {
		t.Fatalf("down from 4 must wrap to 0, got %d", got)
	}
	if got := MoveSteps(0, 0, 1); got != 0 {
		t.Fatalf("empty list must stay 0, got %d", got)
	}
}

func TestEnterLeaveFrame(t *testing.T) {
	if !strings.Contains(EnterFrame(), AltEnter) || !strings.Contains(EnterFrame(), HideCursor) {
		t.Fatal("enter frame must switch screens and hide the cursor")
	}
	if !strings.Contains(LeaveFrame(), AltLeave) || !strings.Contains(LeaveFrame(), ShowCursor) {
		t.Fatal("leave frame must restore the screen and cursor")
	}
}

func TestEnableRestoreRoundTrip(t *testing.T) {
	restore, err := Enable()
	if err != nil {
		t.Skipf("no console in this environment: %v", err)
	}
	restore()
}
