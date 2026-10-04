// Package tui implements the config command interactive menu.
// Stdlib only: line-based menu so it works on PowerShell 5.1 with no
// raw terminal mode. Every view prints its keys in a visible footer.
// Pure renders live in tui_renders.go.
package tui

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/IbatoLionDev/opencode-console-notify/internal/config"
	"github.com/IbatoLionDev/opencode-console-notify/internal/i18n"
)

func readKey(r *bufio.Reader) string {
	line, err := r.ReadString('\n')
	if err != nil && len(line) == 0 {
		return "q"
	}
	return strings.ToLower(strings.TrimSpace(line))
}

const errorFormat = "Error: %s\n"

// session holds the interactive TUI state so each view method stays small.
type session struct {
	pluginsDir string
	settings   config.Settings
	lang       string
	reader     *bufio.Reader
	stdout     io.Writer
}

func isBackKey(key string) bool {
	return key == "q" || key == "esc" || key == "b" || key == "back" || key == "volver"
}

func menuTransition(key string) string {
	switch key {
	case "1", "l", "language", "idioma":
		return "language"
	case "2", "u", "update", "actualizar":
		return "update"
	case "3", "i", "info":
		return "info"
	case "4", "a", "alerts", "alertas":
		return "alerts"
	case "5", "c", "customs", "custom":
		return "customs"
	case "6", "q", "quit", "exit", "salir", "esc":
		return "exit"
	}
	return "menu"
}

// runMenu prints the menu and returns the next view ("exit" quits).
func (t *session) runMenu() string {
	fmt.Fprint(t.stdout, RenderMenu(t.lang))
	return menuTransition(readKey(t.reader))
}

func (t *session) persist() bool {
	if err := config.Save(t.pluginsDir, t.settings); err != nil {
		fmt.Fprintf(t.stdout, errorFormat, err)
		return false
	}
	return true
}

func (t *session) runLanguage() string {
	fmt.Fprint(t.stdout, RenderLanguage(t.lang))
	switch key := readKey(t.reader); key {
	case "1", "en":
		t.lang = "en"
	case "2", "es":
		t.lang = "es"
	default:
		if isBackKey(key) {
			return "menu"
		}
		return "language"
	}
	t.settings.Lang = t.lang
	if !t.persist() {
		return "abort"
	}
	return "language"
}

func (t *session) runAlerts() string {
	fmt.Fprint(t.stdout, RenderAlerts(&t.settings, t.lang))
	key := readKey(t.reader)
	if isBackKey(key) {
		return "menu"
	}
	// Space-prefixed numbers (" 1") also toggle: TrimSpace already
	// removed the space, so the digit alone is enough.
	for _, a := range AllAlertItems(&t.settings, t.lang) {
		if key == a.Key {
			*a.SettingPtr = !*a.SettingPtr
			if !t.persist() {
				return "abort"
			}
		}
	}
	return "alerts"
}

func (t *session) runInfo() string {
	fmt.Fprint(t.stdout, RenderInfo(t.lang))
	readKey(t.reader)
	return "menu"
}

// Run opens the interactive loop. Upgrade runs the existing upgrade flow
// for the current distribution and returns its exit code. Run returns the
// process exit code for the config command.
func Run(pluginsDir string, stdin io.Reader, stdout io.Writer, upgrade func(io.Writer) int) int {
	s, err := config.Load(pluginsDir)
	if err != nil {
		fmt.Fprintf(stdout, errorFormat, err)
		return 1
	}
	t := &session{
		pluginsDir: pluginsDir,
		settings:   s,
		lang:       i18n.Normalize(s.Lang),
		reader:     bufio.NewReader(stdin),
		stdout:     stdout,
	}
	view := "menu"
	for {
		switch view {
		case "menu":
			view = t.runMenu()
		case "language":
			view = t.runLanguage()
		case "alerts":
			view = t.runAlerts()
		case "customs":
			view = t.runCustomsLine()
		case "info":
			view = t.runInfo()
		case "update":
			if upgrade != nil {
				return upgrade(stdout)
			}
			return 0
		case "exit":
			return 0
		default: // "abort": the error is already reported.
			return 1
		}
	}
}
