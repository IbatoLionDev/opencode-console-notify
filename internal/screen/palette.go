// Palette: black background, soft-white body, dark-red selection with
// bright-white text, dim-red borders, light-gray footer. Red is reserved
// for selection and borders so the screen never looks saturated.
package screen

const (
	Reset      = "\x1b[0m"
	Bold       = "\x1b[1m"
	Dim        = "\x1b[2m"
	FgWhite    = "\x1b[97m"
	FgGray     = "\x1b[90m"
	FgRed      = "\x1b[31m"
	BgBlack    = "\x1b[40m"
	BgDarkRed  = "\x1b[48;5;88m"
	BgWhite    = "\x1b[47m"
	HideCursor = "\x1b[?25l"
	ShowCursor = "\x1b[?25h"
	AltEnter   = "\x1b[?1049h"
	AltLeave   = "\x1b[?1049l"
	// MouseEnable turns on basic click + wheel reporting with SGR
	// extended coordinates (DECSET 1000 + 1006). MouseDisable turns
	// both off. Only the fullscreen loop writes them; line mode,
	// pipes and scripts never touch the mouse.
	MouseEnable  = "\x1b[?1000h\x1b[?1006h"
	MouseDisable = "\x1b[?1006l\x1b[?1000l"
	Home         = "\x1b[H"
	Clear        = "\x1b[2J"
	EraseBelow   = "\x1b[J"
)
