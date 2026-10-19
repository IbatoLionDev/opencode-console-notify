//go:build linux

package screen_test

import (
	"testing"

	"github.com/IbatoLionDev/opencode-console-notify/internal/screen"
)

// TestRawTermiosBytesVerbatim asserts the input bytes reach the key
// parser untouched: canonical mode, echo, signals, CR translation and
// software flow control all off.
func TestRawTermiosBytesVerbatim(t *testing.T) {
	old := screen.NewTermios(0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF, 0, make([]uint8, 32), 0, 0)
	got := screen.RawTermios(old)
	if want := uint32(0xFFFF) &^ (screen.IflagBRKINT | screen.IflagICRNL | screen.IflagIXON); screen.Iflag(got) != want {
		t.Fatalf("iflag = %#04x, want %#04x", screen.Iflag(got), want)
	}
	if want := uint32(0xFFFF) &^ (screen.LflagISIG | screen.LflagICANON | screen.LflagECHO | screen.LflagECHONL | screen.LflagIEXTEN); screen.Lflag(got) != want {
		t.Fatalf("lflag = %#04x, want %#04x", screen.Lflag(got), want)
	}
}

// TestRawTermiosKeepsOutputProcessing guards the staircase regression
// on Linux: OPOST (hence ONLCR) must stay on so bare \n keeps reaching
// column 0, exactly like the Windows backend requires.
func TestRawTermiosKeepsOutputProcessing(t *testing.T) {
	old := screen.NewTermios(0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF, 0, make([]uint8, 32), 0, 0)
	got := screen.RawTermios(old)
	if screen.Oflag(got) != screen.Oflag(old) || screen.Cflag(got) != screen.Cflag(old) {
		t.Fatalf("output/control flags must be untouched: got %+v", got)
	}
}
