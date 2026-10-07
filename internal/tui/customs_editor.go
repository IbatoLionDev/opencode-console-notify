// Fullscreen customs editor: create and update flows over user
// notifications. Text input lives in customs_input.go, the event
// picker in customs_picker.go.
package tui

import (
	"fmt"
	"strings"

	"github.com/IbatoLionDev/opencode-console-notify/internal/config"
	"github.com/IbatoLionDev/opencode-console-notify/internal/i18n"
)

// customsTitleKey is the frame title for every custom prompt so the key
// stays in one place instead of repeated across add/edit flows.
const customsTitleKey = "customs.title"

// customsAdd runs create: event picker, then title and body prompts.
func (l *screenLoop) customsAdd() string {
	events := config.KnownEvents()
	idx := l.pickEvent("")
	if idx < 0 {
		return "customs"
	}
	title, ok := l.readLineInput(i18n.T(l.lang, customsTitleKey), i18n.T(l.lang, "customs.titleLabel")+": ")
	if !ok || strings.TrimSpace(title) == "" {
		return "customs"
	}
	body, ok := l.readLineInput(i18n.T(l.lang, customsTitleKey), i18n.T(l.lang, "customs.bodyLabel")+" ("+i18n.T(l.lang, "customs.optional")+"): ")
	if !ok {
		return "customs"
	}
	if _, err := config.AddCustom(&l.settings, events[idx].Type, title, body); err != nil {
		fmt.Fprintf(l.out, errorFormat, err)
		return "abort"
	}
	if !l.persistSettings() {
		return "abort"
	}
	return "customs"
}

// customsEdit re-prompts title and body (empty keeps current) plus the
// event picker preselected on the current event.
func (l *screenLoop) customsEdit(sel int) string {
	if sel < 0 || sel >= len(l.settings.CustomAlerts) {
		return "customs"
	}
	current := l.settings.CustomAlerts[sel]
	title, ok := l.readLineInput(i18n.T(l.lang, customsTitleKey), i18n.T(l.lang, "customs.titleLabel")+" ["+current.Title+"]: ")
	if !ok {
		return "customs"
	}
	if strings.TrimSpace(title) == "" {
		title = current.Title
	}
	body, ok := l.readLineInput(i18n.T(l.lang, customsTitleKey), i18n.T(l.lang, "customs.bodyLabel")+" ["+current.Body+"]: ")
	if !ok {
		return "customs"
	}
	if strings.TrimSpace(body) == "" && current.Body != "" {
		body = current.Body
	}
	idx := l.pickEvent(current.Event)
	if idx < 0 {
		return "customs"
	}
	events := config.KnownEvents()
	kept, err := config.UpdateCustom(&l.settings, current.ID, events[idx].Type, title, body)
	if err != nil {
		fmt.Fprintf(l.out, errorFormat, err)
		return "abort"
	}
	if !kept {
		return "customs"
	}
	if !l.persistSettings() {
		return "abort"
	}
	return "customs"
}
