//go:build linux || darwin

// Package screen testexport exposes identifiers needed by black-box tests
// living in test/internal/screen. Production code must not depend on this
// file; it exists only so the test package can probe internals without
// being in package screen.

package screen

// Termios mirrors the linux/darwin termios struct for black-box tests.
type Termios = termios

// RawTermios exposes rawTermios for linux/darwin tests.
func RawTermios(t termios) termios { return rawTermios(t) }

// NewTermios builds a termios value from exported fields so black-box
// tests can construct one without accessing unexported struct fields.
// cc is copied into the platform-sized cc array (32 bytes on linux,
// 20 bytes on darwin); extra bytes are silently truncated.
func NewTermios(iflag, oflag, cflag, lflag uint32, line uint8, cc []uint8, ispeed, ospeed uint32) termios {
	t := termios{
		iflag:  iflag,
		oflag:  oflag,
		cflag:  cflag,
		lflag:  lflag,
		line:   line,
		ispeed: ispeed,
		ospeed: ospeed,
	}
	copy(t.cc[:], cc)
	return t
}

// Iflag exposes the iflag field of a termios value for black-box tests.
func Iflag(t termios) uint32 { return t.iflag }

// Lflag exposes the lflag field of a termios value for black-box tests.
func Lflag(t termios) uint32 { return t.lflag }

// Oflag exposes the oflag field of a termios value for black-box tests.
func Oflag(t termios) uint32 { return t.oflag }

// Cflag exposes the cflag field of a termios value for black-box tests.
func Cflag(t termios) uint32 { return t.cflag }

// Linux input and local termios flags exposed for black-box tests.
const (
	IflagBRKINT = iflagBRKINT
	IflagICRNL  = iflagICRNL
	IflagIXON   = iflagIXON

	LflagISIG   = lflagISIG
	LflagICANON = lflagICANON
	LflagECHO   = lflagECHO
	LflagECHONL = lflagECHONL
	LflagIEXTEN = lflagIEXTEN
)
