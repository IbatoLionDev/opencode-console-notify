package screen

import (
	"os"
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

// TestEnableKeepsNewlineAutoReturn guards the staircase regression: with
// DISABLE_NEWLINE_AUTO_RETURN (0x0008) set, bare \n drops a row without
// returning to column 0 and every rendered row starts further right. The
// Node stack never touches output modes, so Go must not either.
func TestEnableKeepsNewlineAutoReturn(t *testing.T) {
	restore, err := Enable()
	if err != nil {
		t.Skipf("no console in this environment: %v", err)
	}
	defer restore()
	mode, err := getMode(os.Stdout)
	if err != nil {
		t.Fatalf("cannot read console mode: %v", err)
	}
	if mode&0x0008 != 0 {
		t.Fatal("output mode must keep newline auto-return on (bare \\n must reach column 0)")
	}
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

func TestMouseFrameToggles(t *testing.T) {
	if !strings.Contains(EnterFrame(), MouseEnable) {
		t.Fatal("enter frame must enable SGR mouse")
	}
	if !strings.Contains(LeaveFrame(), MouseDisable) {
		t.Fatal("leave frame must disable mouse")
	}
}
