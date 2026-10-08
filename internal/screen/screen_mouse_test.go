package screen

import (
	"strings"
	"testing"
)

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

func TestParseMouseSGR(t *testing.T) {
	click, size := ParseKey([]byte("\x1b[<0;10;5M"))
	if click.Key != KeyMouseClick || click.MouseX != 10 || click.MouseY != 5 || size != 10 {
		t.Fatalf("left click must parse coords, got %+v size %d", click, size)
	}
	up, _ := ParseKey([]byte("\x1b[<64;1;1M"))
	if up.Key != KeyMouseWheel || up.Wheel != -1 {
		t.Fatalf("wheel up must be -1, got %+v", up)
	}
	down, _ := ParseKey([]byte("\x1b[<65;1;1M"))
	if down.Key != KeyMouseWheel || down.Wheel != 1 {
		t.Fatalf("wheel down must be +1, got %+v", down)
	}
	rel, _ := ParseKey([]byte("\x1b[<0;10;5m"))
	if rel.Key != KeyUnknown {
		t.Fatalf("release must stay unknown so one press never double-fires, got %+v", rel)
	}
	if _, size := ParseKey([]byte("\x1b[<0;10")); size != 0 {
		t.Fatalf("split SGR must wait for more bytes, got size %d", size)
	}
}

func TestItemIndexAtRow(t *testing.T) {
	items := []Item{{Label: "a"}, {Label: "b", Detail: "x"}, {Label: "c"}}
	// Go rows: title=1, border=2, a=3, b=4-5, c=6.
	if idx, ok := ItemIndexAtRow(items, 0, 10, 3); !ok || idx != 0 {
		t.Fatalf("row 3 must hit a, got %d ok=%v", idx, ok)
	}
	if idx, ok := ItemIndexAtRow(items, 0, 10, 5); !ok || idx != 1 {
		t.Fatalf("row 5 must hit b detail, got %d ok=%v", idx, ok)
	}
	if _, ok := ItemIndexAtRow(items, 0, 10, 2); ok {
		t.Fatal("row 2 is the border, must miss")
	}
	if _, ok := ItemIndexAtRow(items, 0, 10, 99); ok {
		t.Fatal("row past the list must miss")
	}
}

func TestParseMouseMotion(t *testing.T) {
	hover, size := ParseKey([]byte("\x1b[<35;12;7M"))
	if hover.Key != KeyMouseMotion || hover.MouseX != 12 || hover.MouseY != 7 || size != 11 {
		t.Fatalf("hover must parse coords, got %+v size %d", hover, size)
	}
	drag, _ := ParseKey([]byte("\x1b[<32;1;1M"))
	if drag.Key != KeyMouseMotion {
		t.Fatalf("drag must report motion, got %+v", drag)
	}
	rel, _ := ParseKey([]byte("\x1b[<35;1;1m"))
	if rel.Key != KeyUnknown {
		t.Fatalf("motion release must stay unknown, got %+v", rel)
	}
	if _, size := ParseKey([]byte("\x1b[<35;12")); size != 0 {
		t.Fatalf("split motion must wait for more bytes, got size %d", size)
	}
}

func TestMouseFrameToggles(t *testing.T) {
	if !strings.Contains(EnterFrame(), MouseEnable) {
		t.Fatal("enter frame must enable SGR mouse")
	}
	if !strings.Contains(LeaveFrame(), MouseDisable) {
		t.Fatal("leave frame must disable mouse")
	}
}
