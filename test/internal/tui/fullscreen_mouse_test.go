package tui_test

import (
	"testing"

	"github.com/IbatoLionDev/opencode-console-notify/internal/config"
	"github.com/IbatoLionDev/opencode-console-notify/internal/tui"
)

// Hover motion rows for height 20: title=1, border=2, items from 3.
func TestMenuHoverMovesSelection(t *testing.T) {
	dir := t.TempDir()
	l, _ := tui.NewTestLoop(dir, "en", "\x1b[<35;1;7M")
	sel := 0
	if view := l.RunMenu(&sel); view != "menu" {
		t.Fatalf("hover must stay on menu, got %q", view)
	}
	if sel != 4 {
		t.Fatalf("hover row 7 must select customs (4), got %d", sel)
	}
}

func TestAlertsHoverMovesWithoutToggling(t *testing.T) {
	dir := t.TempDir()
	l, _ := tui.NewTestLoop(dir, "en", "\x1b[<35;1;5M")
	sel := 0
	if view := l.RunAlerts(&sel); view != "alerts" {
		t.Fatalf("hover must stay on alerts, got %q", view)
	}
	if sel != 1 {
		t.Fatalf("hover row 5 must select the second alert (1), got %d", sel)
	}
	s, err := config.Load(dir)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if !s.Alerts.SessionIdle || !s.Alerts.SessionError {
		t.Fatal("hover must never toggle alerts")
	}
}

func TestLanguageHoverMovesSelection(t *testing.T) {
	dir := t.TempDir()
	l, _ := tui.NewTestLoop(dir, "en", "\x1b[<35;1;4M")
	sel := -1
	if view := l.RunLanguage(&sel); view != "language" {
		t.Fatalf("hover must stay on language, got %q", view)
	}
	if sel != 1 {
		t.Fatalf("hover row 4 must select Spanish (1), got %d", sel)
	}
	if l.Settings.Lang != "en" {
		t.Fatalf("hover must not switch language, got %q", l.Settings.Lang)
	}
}

func TestInfoIgnoresHover(t *testing.T) {
	dir := t.TempDir()
	l, _ := tui.NewTestLoop(dir, "en", "\x1b[<35;1;6M")
	if view := l.RunInfo(); view != "info" {
		t.Fatalf("hover must stay on info, got %q", view)
	}
}
