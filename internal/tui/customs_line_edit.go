// Line-mode customs mutations: event picking plus create, edit and
// delete over user notifications. Renders and dispatch live in
// customs_line.go.
package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/IbatoLionDev/opencode-console-notify/internal/config"
	"github.com/IbatoLionDev/opencode-console-notify/internal/i18n"
)

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
