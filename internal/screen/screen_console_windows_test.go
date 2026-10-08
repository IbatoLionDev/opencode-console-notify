//go:build windows

package screen

import (
	"os"
	"testing"
)

// TestEnableKeepsNewlineAutoReturn guards the staircase regression: with
// DISABLE_NEWLINE_AUTO_RETURN (0x0008) set, bare \n drops a row without
// returning to column 0 and every rendered row starts further right. The
// Node stack never touches output modes, so Go must not either.
// Windows-only: it reads the console output mode via getMode.
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
