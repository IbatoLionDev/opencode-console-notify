// Line-mode customs panel: numbered list with one-line commands.
// Fullscreen twin lives in customs_panel.go; text prompts here read raw
// lines so titles keep their case.
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

// customsLinePickEvent prints the catalog and reads a 1-based number,
// -1 on empty/invalid (back to the panel, never an error).
func (t *session) customsLinePickEvent(current string) int {
	fmt.Fprint(t.stdout, renderCatalogLine(t.lang))
	prompt := i18n.T(t.lang, "customs.pickEvent") + ": "
	if current != "" {
		prompt = i18n.T(t.lang, "customs.pickEvent") + " [" + current + "]: "
	}
	fmt.Fprint(t.stdout, prompt)
	line := readRawLine(t.reader)
	if line == "" {
		if current != "" {
			return catalogIndex(current)
		}
		return -1
	}
	n, err := strconv.Atoi(strings.TrimSpace(line))
	if err != nil || n < 1 || n > len(config.KnownEvents()) {
		return -1
	}
	return n - 1
}

func catalogIndex(event string) int {
	for i, e := range config.KnownEvents() {
		if e.Type == event {
			return i
		}
	}
	return -1
}

func (t *session) customsLineAdd() string {
	idx := t.customsLinePickEvent("")
	if idx < 0 {
		return "customs"
	}
	events := config.KnownEvents()
	fmt.Fprint(t.stdout, i18n.T(t.lang, "customs.titleLabel")+": ")
	title := readRawLine(t.reader)
	if title == "" {
		return "customs"
	}
	fmt.Fprint(t.stdout, i18n.T(t.lang, "customs.bodyLabel")+" ("+i18n.T(t.lang, "customs.optional")+"): ")
	body := readRawLine(t.reader)
	if _, err := config.AddCustom(&t.settings, events[idx].Type, title, body); err != nil {
		fmt.Fprintf(t.stdout, errorFormat, err)
		return "abort"
	}
	if !t.persist() {
		return "abort"
	}
	return "customs"
}

func (t *session) customsLineEdit(idx int) string {
	if idx < 0 || idx >= len(t.settings.CustomAlerts) {
		return "customs"
	}
	current := t.settings.CustomAlerts[idx]
	fmt.Fprint(t.stdout, i18n.T(t.lang, "customs.titleLabel")+" ["+current.Title+"]: ")
	title := readRawLine(t.reader)
	if title == "" {
		title = current.Title
	}
	fmt.Fprint(t.stdout, i18n.T(t.lang, "customs.bodyLabel")+" ["+current.Body+"]: ")
	body := readRawLine(t.reader)
	if body == "" {
		body = current.Body
	}
	eventIdx := t.customsLinePickEvent(current.Event)
	if eventIdx < 0 {
		return "customs"
	}
	events := config.KnownEvents()
	kept, err := config.UpdateCustom(&t.settings, current.ID, events[eventIdx].Type, title, body)
	if err != nil {
		fmt.Fprintf(t.stdout, errorFormat, err)
		return "abort"
	}
	if !kept {
		return "customs"
	}
	if !t.persist() {
		return "abort"
	}
	return "customs"
}

func (t *session) customsLineDelete(idx int) string {
	if idx < 0 || idx >= len(t.settings.CustomAlerts) {
		return "customs"
	}
	title := t.settings.CustomAlerts[idx].Title
	fmt.Fprintf(t.stdout, i18n.T(t.lang, "customs.confirmDelete"), title)
	line := readKey(t.reader)
	if line != "y" && line != "s" {
		return "customs"
	}
	id := t.settings.CustomAlerts[idx].ID
	if !config.RemoveCustom(&t.settings, id) {
		return "customs"
	}
	if !t.persist() {
		return "abort"
	}
	return "customs"
}
