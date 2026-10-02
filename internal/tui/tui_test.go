package tui

import (
	"io"
	"strings"
	"testing"

	"github.com/IbatoLionDev/opencode-console-notify/internal/config"
)

func TestRenderMenuShowsKeys(t *testing.T) {
	for _, lang := range []string{"en", "es"} {
		out := RenderMenu(lang)
		for _, want := range []string{"1.", "2.", "3.", "4.", "5.", "Q"} {
			if !strings.Contains(out, want) {
				t.Fatalf("lang %s menu misses %q in %q", lang, want, out)
			}
		}
	}
}

func TestRenderAlertsListsFourTriggers(t *testing.T) {
	s := config.Defaults()
	for _, lang := range []string{"en", "es"} {
		out := RenderAlerts(&s, lang)
		for _, want := range []string{"session.idle", "session.error", "permission.asked", "question.asked"} {
			if !strings.Contains(out, want) {
				t.Fatalf("lang %s alerts miss %q", lang, want)
			}
		}
	}
}

func TestRenderInfoHasLinks(t *testing.T) {
	out := RenderInfo("en")
	for _, want := range []string{"npm:", "repo:", "releases:", "config"} {
		if !strings.Contains(out, want) {
			t.Fatalf("info misses %q in %q", want, out)
		}
	}
}

func TestInteractiveToggleAndLanguage(t *testing.T) {
	dir := t.TempDir()
	// Alerts: toggle session.idle off (key 1), then back.
	if code := Run(dir, strings.NewReader("4\n1\nq\n5\n"), io.Discard, nil); code != 0 {
		t.Fatalf("alerts run exit = %d", code)
	}
	s, err := config.Load(dir)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if s.Alerts.SessionIdle {
		t.Fatal("sessionIdle should be off after toggle")
	}
	// Language: switch to Spanish (key 2), then back.
	if code := Run(dir, strings.NewReader("1\n2\nq\n5\n"), io.Discard, nil); code != 0 {
		t.Fatalf("language run exit = %d", code)
	}
	s, err = config.Load(dir)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if s.Lang != "es" {
		t.Fatalf("lang = %q, want es", s.Lang)
	}
}
