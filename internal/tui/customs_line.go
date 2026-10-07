// Line-mode customs panel: numbered list with one-line commands.
// Fullscreen twin lives in customs_panel.go; text prompts here read raw
// lines so titles keep their case. Mutations (pick/add/edit/delete)
// live in customs_line_edit.go.
package tui

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"github.com/IbatoLionDev/opencode-console-notify/internal/config"
	"github.com/IbatoLionDev/opencode-console-notify/internal/i18n"
)

// readRawLine reads one line trimmed but case-preserved (titles must not
// be lowercased). EOF yields "".
func readRawLine(r *bufio.Reader) string {
	line, err := r.ReadString('\n')
	if err != nil && len(line) == 0 {
		return ""
	}
	return strings.TrimSpace(line)
}

// renderCustomsLine numbers customs with states; empty shows the hint.
func renderCustomsLine(s *config.Settings, lang string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", i18n.T(lang, "customs.title"))
	if len(s.CustomAlerts) == 0 {
		fmt.Fprintf(&b, "%s\n", i18n.T(lang, "customs.empty"))
	}
	for i, c := range s.CustomAlerts {
		fmt.Fprintf(&b, "%d. %s [%s] [%s]\n   %s\n", i+1, c.Title, i18n.T(lang, "custom.tag"), i18n.T(lang, boolStateKey(c.Enabled)), customLineDetail(c))
	}
	fmt.Fprintf(&b, "%s\n", i18n.T(lang, "customs.lineFooter"))
	return b.String()
}

func customLineDetail(c config.CustomAlert) string {
	if c.Body == "" {
		return c.Event
	}
	return c.Event + " — " + c.Body
}

// renderCatalogLine numbers every catalog event with its description,
// marking noisy ones.
func renderCatalogLine(lang string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s\n", i18n.T(lang, "customs.catalogTitle"))
	for i, e := range config.KnownEvents() {
		desc := e.DescEn
		if lang == "es" {
			desc = e.DescEs
		}
		mark := ""
		if e.Noisy {
			mark = " [!]"
		}
		fmt.Fprintf(&b, "%d. %s: %s%s\n", i+1, e.Type, desc, mark)
	}
	fmt.Fprintf(&b, "%s\n", i18n.T(lang, "footer.back"))
	return b.String()
}

// customIndex parses a 1-based position, -1 when outside the list.
func customIndex(s *config.Settings, arg string) int {
	n, err := strconv.Atoi(strings.TrimSpace(arg))
	if err != nil || n < 1 || n > len(s.CustomAlerts) {
		return -1
	}
	return n - 1
}

func (t *session) runCustomsLine() string {
	for {
		fmt.Fprint(t.stdout, renderCustomsLine(&t.settings, t.lang))
		if next := t.dispatchCustomsLine(readKey(t.reader)); next != "customs" {
			return next
		}
	}
}

// dispatchCustomsLine routes one panel command; "customs" means stay.
func (t *session) dispatchCustomsLine(line string) string {
	switch {
	case isBackKey(line):
		return "menu"
	case line == "a":
		return t.customsLineAdd()
	case line == "c":
		fmt.Fprint(t.stdout, renderCatalogLine(t.lang))
		readKey(t.reader)
		return "customs"
	case strings.HasPrefix(line, "e "):
		return t.customsLineEdit(customIndex(&t.settings, strings.TrimPrefix(line, "e ")))
	case strings.HasPrefix(line, "d "):
		return t.customsLineDelete(customIndex(&t.settings, strings.TrimPrefix(line, "d ")))
	default:
		return t.toggleCustomLine(line)
	}
}

// toggleCustomLine flips the custom at a 1-based position; unknown input
// stays on the panel, save failures abort.
func (t *session) toggleCustomLine(line string) string {
	idx := customIndex(&t.settings, line)
	if idx < 0 {
		return "customs"
	}
	id := t.settings.CustomAlerts[idx].ID
	if !config.ToggleCustom(&t.settings, id) {
		return "customs"
	}
	if !t.persist() {
		return "abort"
	}
	return "customs"
}
