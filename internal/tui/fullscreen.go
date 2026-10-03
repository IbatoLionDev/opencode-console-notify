// Fullscreen config loop: the four views rendered in place on the
// alternate screen with keyboard navigation. Line mode (tui.go Run)
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

// screenLoop holds the fullscreen state. Width/height are injected so
// tests drive the loop with scripted bytes.
type screenLoop struct {
	pluginsDir string
	settings   config.Settings
	lang       string
	in         io.Reader
	out        io.Writer
	width      int
	height     int
	upgrade    func(io.Writer) int
	pending    []byte
}

func (l *screenLoop) readKey() (screen.ParsedKey, error) {
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
		n, err := l.in.Read(chunk[:])
		if n > 0 {
			l.pending = append(l.pending, chunk[:n]...)
			continue
		}
		if err != nil {
			return screen.ParsedKey{Key: screen.KeyQuit}, err
		}
	}
}

func menuScreenItems(lang string) []screen.Item {
	return []screen.Item{
		{Label: i18n.T(lang, "menu.language")},
		{Label: i18n.T(lang, "menu.update")},
		{Label: i18n.T(lang, "menu.info")},
		{Label: i18n.T(lang, "menu.alerts")},
		{Label: i18n.T(lang, "menu.exit")},
	}
}

func menuTarget(sel int) string {
	switch sel {
	case 0:
		return "language"
	case 1:
		return "update"
	case 2:
		return "info"
	case 3:
		return "alerts"
	}
	return "exit"
}

func menuShortcut(r rune) string {
	switch r {
	case '1', 'l', 'L':
		return "language"
	case '2', 'u', 'U':
		return "update"
	case '3', 'i', 'I':
		return "info"
	case '4', 'a', 'A':
		return "alerts"
	case '5':
		return "exit"
	}
	return ""
}

func (l *screenLoop) runMenu(sel *int) string {
	items := menuScreenItems(l.lang)
	*sel = screen.ClampSelection(len(items), *sel)
	frame := screen.Frame{
		Title:    i18n.T(l.lang, "app.title"),
		Items:    items,
		Selected: *sel,
		Footer:   i18n.T(l.lang, "footer.menu"),
		Height:   screen.FrameHeight(l.height),
	}
	fmt.Fprint(l.out, screen.Render(frame, l.width))
	k, err := l.readKey()
	if err != nil {
		return "abort"
	}
	if k.Key == screen.KeyRune {
		if target := menuShortcut(k.Rune); target != "" {
			return target
		}
		return "menu"
	}
	switch k.Key {
	case screen.KeyUp:
		*sel = screen.MoveSteps(len(items), *sel, -1)
	case screen.KeyDown:
		*sel = screen.MoveSteps(len(items), *sel, 1)
	case screen.KeyEnter:
		return menuTarget(*sel)
	case screen.KeyQuit, screen.KeyEsc:
		return "exit"
	}
	return "menu"
}

func alertScreenItems(s *config.Settings, lang string) []screen.Item {
	fields := []struct {
		on      bool
		nameKey string
		trigKey string
	}{
		{s.Alerts.SessionIdle, "alert.sessionIdle", "trigger.sessionIdle"},
		{s.Alerts.SessionError, "alert.sessionError", "trigger.sessionError"},
		{s.Alerts.PermissionAsked, "alert.permissionAsked", "trigger.permissionAsked"},
		{s.Alerts.QuestionAsked, "alert.questionAsked", "trigger.questionAsked"},
	}
	items := make([]screen.Item, 0, len(fields))
	for _, f := range fields {
		items = append(items, screen.Item{
			Label:  i18n.T(lang, f.nameKey),
			Detail: i18n.T(lang, f.trigKey),
			State:  i18n.T(lang, boolStateKey(f.on)),
		})
	}
	return items
}

func boolStateKey(on bool) string {
	if on {
		return "state.on"
	}
	return "state.off"
}

func alertSetting(s *config.Settings, idx int) *bool {
	switch idx {
	case 0:
		return &s.Alerts.SessionIdle
	case 1:
		return &s.Alerts.SessionError
	case 2:
		return &s.Alerts.PermissionAsked
	}
	return &s.Alerts.QuestionAsked
}

func (l *screenLoop) persistSettings() bool {
	if err := config.Save(l.pluginsDir, l.settings); err != nil {
		fmt.Fprintf(l.out, "Error: %s\n", err)
		return false
	}
	return true
}

func (l *screenLoop) toggleAlert(idx int) bool {
	*alertSetting(&l.settings, idx) = !*alertSetting(&l.settings, idx)
	return l.persistSettings()
}

func (l *screenLoop) runAlerts(sel *int) string {
	items := alertScreenItems(&l.settings, l.lang)
	*sel = screen.ClampSelection(len(items), *sel)
	frame := screen.Frame{
		Title:    i18n.T(l.lang, "alerts.title"),
		Items:    items,
		Selected: *sel,
		Footer:   i18n.T(l.lang, "footer.toggle"),
		Height:   screen.FrameHeight(l.height),
	}
	fmt.Fprint(l.out, screen.Render(frame, l.width))
	k, err := l.readKey()
	if err != nil {
		return "abort"
	}
	if k.Key == screen.KeyRune {
		if idx := digitIndex(k.Rune, len(items)); idx >= 0 {
			if !l.toggleAlert(idx) {
				return "abort"
			}
		}
		return "alerts"
	}
	switch k.Key {
	case screen.KeyUp:
		*sel = screen.MoveSteps(len(items), *sel, -1)
	case screen.KeyDown:
		*sel = screen.MoveSteps(len(items), *sel, 1)
	case screen.KeyEnter, screen.KeySpace:
		if !l.toggleAlert(*sel) {
			return "abort"
		}
	case screen.KeyQuit, screen.KeyEsc:
		return "menu"
	}
	return "alerts"
}

// digitIndex maps 1-based digit runes to item indexes, -1 when outside.
func digitIndex(r rune, n int) int {
	if r < '1' || int(r-'1') >= n {
		return -1
	}
	return int(r - '1')
}

func (l *screenLoop) runLanguage() string {
	items := []screen.Item{
		{Label: "English", State: currentMark(l.lang == "en", l.lang)},
		{Label: "Español", State: currentMark(l.lang == "es", l.lang)},
	}
	frame := screen.Frame{
		Title:    i18n.T(l.lang, "lang.title"),
		Items:    items,
		Selected: screen.ClampSelection(len(items), langIndex(l.lang)),
		Footer:   i18n.T(l.lang, "footer.back"),
		Height:   screen.FrameHeight(l.height),
	}
	fmt.Fprint(l.out, screen.Render(frame, l.width))
	k, err := l.readKey()
	if err != nil {
		return "abort"
	}
	if k.Key == screen.KeyRune {
		if k.Rune == '1' || k.Rune == 'e' || k.Rune == 'E' {
			return l.setLanguage("en")
		}
		if k.Rune == '2' {
			return l.setLanguage("es")
		}
	}
	switch k.Key {
	case screen.KeyEnter, screen.KeySpace:
		if langIndex(l.lang) == 0 {
			return l.setLanguage("es")
		}
		return l.setLanguage("en")
	case screen.KeyQuit, screen.KeyEsc:
		return "menu"
	}
	return "language"
}

func langIndex(lang string) int {
	if lang == "es" {
		return 1
	}
	return 0
}

func currentMark(active bool, lang string) string {
	if active {
		return i18n.T(lang, "lang.current")
	}
	return ""
}

func (l *screenLoop) setLanguage(lang string) string {
	l.lang = lang
	l.settings.Lang = lang
	if !l.persistSettings() {
		return "abort"
	}
	return "language"
}

func (l *screenLoop) runInfo() string {
	frame := screen.Frame{
		Title:    i18n.T(l.lang, "menu.info"),
		Items:    infoScreenItems(),
		Selected: -1,
		Footer:   i18n.T(l.lang, "footer.back"),
		Height:   screen.FrameHeight(l.height),
	}
	fmt.Fprint(l.out, screen.Render(frame, l.width))
	if _, err := l.readKey(); err != nil {
		return "abort"
	}
	return "menu"
}

func infoScreenItems() []screen.Item {
	return []screen.Item{
		{Label: "install, uninstall, doctor, test, upgrade, version, config"},
		{Label: "npm: " + NPMPage},
		{Label: "repo: " + RepoURL},
		{Label: "releases: " + ReleasesURL},
	}
}

// doUpdate leaves the alternate screen first so upgrade output stays
// visible, then runs the existing upgrade flow for this distribution.
func (l *screenLoop) doUpdate(restore func()) int {
	fmt.Fprint(l.out, screen.LeaveFrame())
	restore()
	if l.upgrade != nil {
		return l.upgrade(l.out)
	}
	return 0
}

func (l *screenLoop) loop(restore func()) int {
	menuSel := 0
	alertSel := 0
	view := "menu"
	for {
		switch view {
		case "menu":
			view = l.runMenu(&menuSel)
		case "alerts":
			view = l.runAlerts(&alertSel)
		case "language":
			view = l.runLanguage()
		case "info":
			view = l.runInfo()
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
		fmt.Fprintf(stdout, "Error: %s\n", err)
		return 1
	}
	restore, err := screen.Enable()
	if err != nil {
		fmt.Fprintf(stdout, "Error: %s\n", err)
		return 1
	}
	defer restore()
	fmt.Fprint(stdout, screen.EnterFrame())
	w, h := screen.Size()
	l := &screenLoop{
		pluginsDir: pluginsDir,
		settings:   s,
		lang:       i18n.Normalize(s.Lang),
		in:         stdin,
		out:        stdout,
		width:      w,
		height:     h,
		upgrade:    upgrade,
	}
	code := l.loop(restore)
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
