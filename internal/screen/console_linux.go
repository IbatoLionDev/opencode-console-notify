//go:build linux

package screen

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

// Console setup for Linux: termios raw input plus window size via
// ioctl, Go stdlib only (no x/sys dependency). Field order and sizes
// match golang.org/x/sys/unix.Termios exactly (60 bytes); the ioctl
// numbers below hold on amd64 and arm64.
const (
	tcgets     = 0x5401
	tcsets     = 0x5402
	tiocgwinsz = 0x5413

	iflagBRKINT = 0x0002
	iflagICRNL  = 0x0100
	iflagIXON   = 0x0400

	lflagISIG   = 0x0001
	lflagICANON = 0x0002
	lflagECHO   = 0x0008
	lflagECHONL = 0x0040
	lflagIEXTEN = 0x8000
)

// termios mirrors the Linux struct termios for TCGETS/TCSETS. Only the
// input and local flags are ever changed; output processing (OPOST,
// hence ONLCR) stays on so bare \n keeps reaching column 0 exactly
// like the Windows backend requires.
type termios struct {
	iflag  uint32
	oflag  uint32
	cflag  uint32
	lflag  uint32
	line   uint8
	cc     [32]uint8
	ispeed uint32
	ospeed uint32
}

// winsize mirrors the Linux struct winsize for TIOCGWINSZ.
type winsize struct {
	rows   uint16
	cols   uint16
	xpixel uint16
	ypixel uint16
}

// rawTermios returns t with canonical mode, echo and signals off and
// input bytes verbatim (no CR translation, no flow control). Output
// flags are deliberately untouched.
func rawTermios(t termios) termios {
	t.iflag &^= iflagBRKINT | iflagICRNL | iflagIXON
	t.lflag &^= lflagISIG | lflagICANON | lflagECHO | lflagECHONL | lflagIEXTEN
	return t
}

func ioctl(fd, req uintptr, arg unsafe.Pointer) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, req, uintptr(arg))
	if errno != 0 {
		return errno
	}
	return nil
}

func getTermios(f *os.File) (termios, error) {
	var t termios
	if err := ioctl(f.Fd(), tcgets, unsafe.Pointer(&t)); err != nil {
		return t, fmt.Errorf("get terminal attributes: %w", err)
	}
	return t, nil
}

func setTermios(f *os.File, t *termios) error {
	if err := ioctl(f.Fd(), tcsets, unsafe.Pointer(t)); err != nil {
		return fmt.Errorf("set terminal attributes: %w", err)
	}
	return nil
}

// Size reports the terminal size in columns and rows, defaulting to
// 80x24 when the terminal cannot answer.
func Size() (int, int) {
	var ws winsize
	if err := ioctl(os.Stdout.Fd(), tiocgwinsz, unsafe.Pointer(&ws)); err != nil {
		return 80, 24
	}
	w, h := int(ws.cols), int(ws.rows)
	if w < 10 {
		w = 10
	}
	if h < 4 {
		h = 4
	}
	return w, h
}

// Enable prepares the console for fullscreen input: raw termios on
// stdin (verbatim key bytes, SGR mouse reports included) while output
// processing stays on for rendering. It returns a restore function
// that must run on every exit path, including errors. Without a
// terminal it fails honestly so callers fall back to line mode.
func Enable() (restore func(), err error) {
	old, err := getTermios(os.Stdin)
	if err != nil {
		return nil, err
	}
	raw := rawTermios(old)
	if err := setTermios(os.Stdin, &raw); err != nil {
		return nil, err
	}
	return func() {
		setTermios(os.Stdin, &old)
	}, nil
}
