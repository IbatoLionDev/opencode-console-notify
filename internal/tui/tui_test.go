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
		for _, want := range []string{"1.", "2.", "3.", "4.", "5.", "6.", "Q"} {
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
	if code := Run(dir, strings.NewReader("4\n1\nq\n6\n"), io.Discard, nil); code != 0 {
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
	if code := Run(dir, strings.NewReader("1\n2\nq\n6\n"), io.Discard, nil); code != 0 {
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

func TestLineCustomsLifecycle(t *testing.T) {
	dir := t.TempDir()
	// Menu 5 customs, A add, event 1, title, body, back, exit.
	if code := Run(dir, strings.NewReader("5\na\n1\nLine Title\nLine Body\nq\nq\n"), io.Discard, nil); code != 0 {
		t.Fatalf("create run exit = %d", code)
	}
	s, err := config.Load(dir)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(s.CustomAlerts) != 1 || s.CustomAlerts[0].Event != "session.created" || s.CustomAlerts[0].Title != "Line Title" {
		t.Fatalf("unexpected customs: %+v", s.CustomAlerts)
	}
	// Toggle #1 off, then delete it with y confirm.
	if code := Run(dir, strings.NewReader("5\n1\nq\nq\n"), io.Discard, nil); code != 0 {
		t.Fatalf("toggle run exit = %d", code)
	}
	s, err = config.Load(dir)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if s.CustomAlerts[0].Enabled {
		t.Fatal("custom must be off after toggle")
	}
	if code := Run(dir, strings.NewReader("5\nd 1\ny\nq\nq\n"), io.Discard, nil); code != 0 {
		t.Fatalf("delete run exit = %d", code)
	}
	s, err = config.Load(dir)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(s.CustomAlerts) != 0 {
		t.Fatalf("custom must be deleted, got %+v", s.CustomAlerts)
	}
}
