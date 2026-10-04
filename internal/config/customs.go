// Custom alerts: user text bound to catalog OpenCode events. The user
// never alters events, only picks one. Validation is catalog-strict;
// titles and bodies are rune-truncated to toast-safe length.
package config

import (
	"fmt"
	"strings"
)

// CustomTextLimit caps titles and bodies in runes (matches MAX_LINE).
const CustomTextLimit = 160

// CustomAlert is one user notification: text plus the subscribed event.
type CustomAlert struct {
	ID      string `json:"id"`
	Event   string `json:"event"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	Enabled bool   `json:"enabled"`
}

// ValidateCustom rejects unknown events and empty titles.
func ValidateCustom(event, title string) error {
	if !IsKnownEvent(event) {
		return fmt.Errorf("unknown event %q (see --list-events)", event)
	}
	if strings.TrimSpace(title) == "" {
		return fmt.Errorf("custom title must not be empty")
	}
	return nil
}

// TruncateCustomText cuts s to CustomTextLimit runes without splitting
// multibyte sequences.
func TruncateCustomText(s string) string {
	runes := []rune(s)
	if len(runes) <= CustomTextLimit {
		return s
	}
	return string(runes[:CustomTextLimit])
}

// NextCustomID allocates the next stable id (custom-1, custom-2, ...),
// skipping ids still in use so deletes never collide.
func NextCustomID(s Settings) string {
	max := 0
	for _, c := range s.CustomAlerts {
		var n int
		if _, err := fmt.Sscanf(c.ID, "custom-%d", &n); err == nil && n > max {
			max = n
		}
	}
	return fmt.Sprintf("custom-%d", max+1)
}

// FindCustom returns the custom with id, or nil.
func FindCustom(s Settings, id string) *CustomAlert {
	for i := range s.CustomAlerts {
		if s.CustomAlerts[i].ID == id {
			return &s.CustomAlerts[i]
		}
	}
	return nil
}

// AddCustom validates and appends an enabled custom, returning its id.
func AddCustom(s *Settings, event, title, body string) (string, error) {
	if err := ValidateCustom(event, title); err != nil {
		return "", err
	}
	id := NextCustomID(*s)
	s.CustomAlerts = append(s.CustomAlerts, CustomAlert{
		ID:      id,
		Event:   strings.TrimSpace(event),
		Title:   TruncateCustomText(strings.TrimSpace(title)),
		Body:    TruncateCustomText(strings.TrimSpace(body)),
		Enabled: true,
	})
	return id, nil
}

// UpdateCustom replaces event/title/body of id (validated, stays enabled
// as it was). It reports whether id existed.
func UpdateCustom(s *Settings, id, event, title, body string) (bool, error) {
	if err := ValidateCustom(event, title); err != nil {
		return false, err
	}
	c := FindCustom(*s, id)
	if c == nil {
		return false, nil
	}
	c.Event = strings.TrimSpace(event)
	c.Title = TruncateCustomText(strings.TrimSpace(title))
	c.Body = TruncateCustomText(strings.TrimSpace(body))
	return true, nil
}

// RemoveCustom deletes id, reporting whether it existed.
func RemoveCustom(s *Settings, id string) bool {
	for i := range s.CustomAlerts {
		if s.CustomAlerts[i].ID != id {
			continue
		}
		s.CustomAlerts = append(s.CustomAlerts[:i], s.CustomAlerts[i+1:]...)
		return true
	}
	return false
}

// ToggleCustom flips enabled of id, reporting whether it existed.
func ToggleCustom(s *Settings, id string) bool {
	c := FindCustom(*s, id)
	if c == nil {
		return false
	}
	c.Enabled = !c.Enabled
	return true
}

// CustomsForEvent returns the enabled customs bound to an event type,
// in definition order.
func CustomsForEvent(s Settings, event string) []CustomAlert {
	var out []CustomAlert
	for _, c := range s.CustomAlerts {
		if c.Enabled && c.Event == event {
			out = append(out, c)
		}
	}
	return out
}
