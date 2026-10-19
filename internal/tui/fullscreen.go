// Fullscreen config loop: keyboard navigation on the alternate screen.
// The four views live in fullscreen_views.go; line mode (tui.go Run)
// stays as the fallback for pipes and scripts.
package tui

import (
	"fmt"
	"io"
	"os"

	"github.com/IbatoLionDev/opencode-console-notify/internal/config"
	"github.com/IbatoLionDev/opencode-console-notify/internal/i18n"
	"github.com/IbatoLionDev/opencode-console-notify/internal/screen"
)

// ScreenLoop holds the fullscreen state. Width/height are injected so
// tests drive the loop with scripted bytes.
type ScreenLoop struct {
	PluginsDir string
	Settings   config.Settings
	Lang       string
	In         io.Reader
	Out        io.Writer
	Width      int
	Height     int
	Upgrade    func(io.Writer) int
	pending    []byte
}

func (l *ScreenLoop) readKey() (screen.ParsedKey, error) {
	for {
		if len(l.pending) > 0 {
			k, n := screen.ParseKey(l.pending)
			if n > 0 {
				l.pending = l.pending[n:]
				return k, nil
			}
			if len(l.pending) >= 8 {
				l.pending = l.pending[1:]
				return screen.ParsedKey{Key: screen.KeyUnknown}, nil
			}
		}
		var chunk [16]byte
		n, err := l.In.Read(chunk[:])
		if n > 0 {
			l.pending = append(l.pending, chunk[:n]...)
			continue
		}
		if err != nil {
			return screen.ParsedKey{Key: screen.KeyQuit}, err
		}
	}
}

// persistSettings saves the loop settings, reporting failures on the
// screen; every view shares it.
func (l *ScreenLoop) persistSettings() bool {
	if err := config.Save(l.PluginsDir, l.Settings); err != nil {
		fmt.Fprintf(l.Out, errorFormat, err)
		return false
	}
	return true
}

// doUpdate leaves the alternate screen first so upgrade output stays
// visible, then runs the existing upgrade flow for this distribution.
func (l *ScreenLoop) doUpdate(restore func()) int {
	fmt.Fprint(l.Out, screen.LeaveFrame())
	restore()
	if l.Upgrade != nil {
		return l.Upgrade(l.Out)
	}
	return 0
}

func (l *ScreenLoop) Loop(restore func()) int {
	menuSel := 0
	alertSel := 0
	customsSel := 0
	languageSel := -1
	view := "menu"
	for {
		switch view {
		case "menu":
			view = l.RunMenu(&menuSel)
		case "alerts":
			view = l.RunAlerts(&alertSel)
		case "customs":
			view = l.RunCustoms(&customsSel)
		case "language":
			view = l.RunLanguage(&languageSel)
		case "info":
			view = l.RunInfo()
		case "update":
			return l.doUpdate(restore)
		case "exit":
			return 0
		default: // "abort": the error is already reported.
			return 1
		}
	}
}

// RunFullscreen opens the alternate screen and runs the keyboard loop.
// The caller falls back to line mode when the console is unavailable.
func RunFullscreen(pluginsDir string, stdin io.Reader, stdout io.Writer, upgrade func(io.Writer) int) int {
	s, err := config.Load(pluginsDir)
	if err != nil {
		fmt.Fprintf(stdout, errorFormat, err)
		return 1
	}
	restore, err := screen.Enable()
	if err != nil {
		fmt.Fprintf(stdout, errorFormat, err)
		return 1
	}
	defer restore()
	fmt.Fprint(stdout, screen.EnterFrame())
	w, h := screen.Size()
	l := &ScreenLoop{
		PluginsDir: pluginsDir,
		Settings:   s,
		Lang:       i18n.Normalize(s.Lang),
		In:         stdin,
		Out:        stdout,
		Width:      w,
		Height:     h,
		Upgrade:    upgrade,
	}
	code := l.Loop(restore)
	fmt.Fprint(stdout, screen.LeaveFrame())
	return code
}

// RunSmart opens the fullscreen loop on a real console and the line mode
// everywhere else (pipes, scripts, probes). Stdin must stay open for the
// console check, so os.Stdin callers pass it directly.
func RunSmart(pluginsDir string, stdin *os.File, stdout io.Writer, upgrade func(io.Writer) int) int {
	if isConsole(stdin) {
		return RunFullscreen(pluginsDir, stdin, stdout, upgrade)
	}
	return Run(pluginsDir, stdin, stdout, upgrade)
}

// isConsole reports whether f is a character device (a real console).
func isConsole(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
