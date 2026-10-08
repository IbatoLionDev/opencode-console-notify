//go:build linux

package screen

import (
	"testing"
)

// TestRawTermiosBytesVerbatim asserts the input bytes reach the key
// parser untouched: canonical mode, echo, signals, CR translation and
// software flow control all off.
func TestRawTermiosBytesVerbatim(t *testing.T) {
	old := termios{iflag: 0xFFFF, lflag: 0xFFFF}
	got := rawTermios(old)
	if want := uint32(0xFFFF) &^ (iflagBRKINT | iflagICRNL | iflagIXON); got.iflag != want {
		t.Fatalf("iflag = %#04x, want %#04x", got.iflag, want)
	}
	if want := uint32(0xFFFF) &^ (lflagISIG | lflagICANON | lflagECHO | lflagECHONL | lflagIEXTEN); got.lflag != want {
		t.Fatalf("lflag = %#04x, want %#04x", got.lflag, want)
	}
}

// TestRawTermiosKeepsOutputProcessing guards the staircase regression
// on Linux: OPOST (hence ONLCR) must stay on so bare \n keeps reaching
// column 0, exactly like the Windows backend requires.
func TestRawTermiosKeepsOutputProcessing(t *testing.T) {
	old := termios{oflag: 0xFFFF, cflag: 0xFFFF}
	got := rawTermios(old)
	if got.oflag != old.oflag || got.cflag != old.cflag {
		t.Fatalf("output/control flags must be untouched: got %+v", got)
	}
}
