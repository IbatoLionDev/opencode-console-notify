//go:build windows

package screen

import (
	"errors"
	"os"
	"syscall"
	"unsafe"
)

// Virtual-terminal and input flags (kernel32 console modes). Defined here
// so the screen package stays on Go stdlib with no x/sys dependency.
const (
	enableProcessedInput = 0x0001
	enableLineInput      = 0x0002
	enableEchoInput      = 0x0004
	enableWindowInput    = 0x0008
	enableVTInput        = 0x0200
	enableVTProcessing   = 0x0004
)

// kernel32 console procs used directly because stdlib syscall only exposes
// the Get half of the console-mode pair.
var procSetConsoleMode = syscall.NewLazyDLL("kernel32.dll").NewProc("SetConsoleMode")

var procGetScreenInfo = syscall.NewLazyDLL("kernel32.dll").NewProc("GetConsoleScreenBufferInfo")

// windowRect is the visible console window in character cells.
type windowRect struct {
	left   int16
	top    int16
	right  int16
	bottom int16
}

// screenBufferInfo mirrors CONSOLE_SCREEN_BUFFER_INFO up to the window
// rect; later fields are not needed.
type screenBufferInfo struct {
	size   [2]int16
	cursor [2]int16
	attrs  uint16
	window windowRect
	maxWin [2]int16
}

// Size reports the visible console window in columns and rows, defaulting
// to 80x24 when the console cannot answer.
func Size() (int, int) {
	var info screenBufferInfo
	r1, _, _ := procGetScreenInfo.Call(uintptr(syscall.Handle(os.Stdout.Fd())), uintptr(unsafe.Pointer(&info)))
	if r1 == 0 {
		return 80, 24
	}
	w := int(info.window.right-info.window.left) + 1
	h := int(info.window.bottom-info.window.top) + 1
	if w < 10 {
		w = 10
	}
	if h < 4 {
		h = 4
	}
	return w, h
}

// Enable prepares the console for fullscreen input and ANSI output: raw
// key bytes (line and echo off, VT input on so arrows arrive as escape
// sequences) and VT processing on stdout. It returns a restore function
// that must run on every exit path, including errors.
func Enable() (restore func(), err error) {
	inModes, err := getMode(os.Stdin)
	if err != nil {
		return nil, err
	}
	outModes, err := getMode(os.Stdout)
	if err != nil {
		return nil, err
	}
	in := inModes &^ (enableLineInput | enableEchoInput | enableProcessedInput)
	in |= enableVTInput | enableWindowInput
	if err := setMode(os.Stdin, in); err != nil {
		return nil, err
	}
	// Keep newline auto-return ON (never set DISABLE_NEWLINE_AUTO_RETURN):
	// Render terminates rows with bare \n exactly like the Node stack, so
	// \n must move to column 0 of the next row. Disabling auto-return
	// produces the staircase effect (every row starts further right).
	out := outModes | enableVTProcessing
	if err := setMode(os.Stdout, out); err != nil {
		setMode(os.Stdin, inModes)
		return nil, err
	}
	return func() {
		setMode(os.Stdout, outModes)
		setMode(os.Stdin, inModes)
	}, nil
}

func getMode(f *os.File) (uint32, error) {
	var mode uint32
	if err := syscall.GetConsoleMode(syscall.Handle(f.Fd()), &mode); err != nil {
		return 0, err
	}
	return mode, nil
}

func setMode(f *os.File, mode uint32) error {
	// SetConsoleMode is not exposed by Go's stdlib syscall package, so it
	// goes through kernel32 directly (still stdlib, no x/sys dependency).
	r1, _, err := procSetConsoleMode.Call(uintptr(syscall.Handle(f.Fd())), uintptr(mode))
	if r1 == 0 {
		if err != nil && err != syscall.Errno(0) {
			return err
		}
		return errors.New("SetConsoleMode failed")
	}
	return nil
}
