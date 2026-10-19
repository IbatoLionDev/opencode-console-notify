package tui_test

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/IbatoLionDev/opencode-console-notify/internal/config"
	"github.com/IbatoLionDev/opencode-console-notify/internal/tui"
)

func TestFullscreenNavigateExit(t *testing.T) {
	dir := t.TempDir()
	l, out := tui.NewTestLoop(dir, "en", "j\x1b[Aq")
	if code := l.Loop(tui.NoopRestore); code != 0 {
		t.Fatalf(tui.ExitCodeFormat, code)
	}
	if !strings.Contains(out.String(), "> ") {
		t.Fatal("output must mark the selection in place")
	}
}

func TestFullscreenTogglePersists(t *testing.T) {
	dir := t.TempDir()
	l, _ := tui.NewTestLoop(dir, "en", "4 qq")
	if code := l.Loop(tui.NoopRestore); code != 0 {
		t.Fatalf(tui.ExitCodeFormat, code)
	}
	s, err := config.Load(dir)
	if err != nil {
		t.Fatalf(tui.LoadFailedFmt, err)
	}
	if s.Alerts.SessionIdle {
		t.Fatal("space must toggle sessionIdle off and persist it")
	}
}

func TestFullscreenLanguagePersists(t *testing.T) {
	dir := t.TempDir()
	l, _ := tui.NewTestLoop(dir, "en", "12qq")
	if code := l.Loop(tui.NoopRestore); code != 0 {
		t.Fatalf(tui.ExitCodeFormat, code)
	}
	s, err := config.Load(dir)
	if err != nil {
		t.Fatalf(tui.LoadFailedFmt, err)
	}
	if s.Lang != "es" {
		t.Fatalf("lang = %q, want es", s.Lang)
	}
}

func TestFullscreenInfoBacksOut(t *testing.T) {
	dir := t.TempDir()
	l, out := tui.NewTestLoop(dir, "en", "3xq")
	if code := l.Loop(tui.NoopRestore); code != 0 {
		t.Fatalf(tui.ExitCodeFormat, code)
	}
	if !strings.Contains(out.String(), "releases:") {
		t.Fatal("info view must render the links in place")
	}
}

func TestFullscreenUpdateDelegates(t *testing.T) {
	dir := t.TempDir()
	l, _ := tui.NewTestLoop(dir, "en", "2")
	called := false
	l.Upgrade = func(w io.Writer) int {
		called = true
		return 7
	}
	if code := l.Loop(tui.NoopRestore); code != 7 {
		t.Fatalf("exit code = %d, want upgrade code 7", code)
	}
	if !called {
		t.Fatal("update view must run the upgrade flow")
	}
}

func TestMenuTarget(t *testing.T) {
	for sel, want := range map[int]string{0: "language", 1: "update", 2: "info", 3: "alerts", 4: "customs", 5: "exit"} {
		if got := tui.MenuTarget(sel); got != want {
			t.Fatalf("menuTarget(%d) = %q, want %q", sel, got, want)
		}
	}
}

func TestFullscreenCustomCreateToggle(t *testing.T) {
	dir := t.TempDir()
	// Menu 5 customs, A add, Enter picks first event, Hi + empty body, Space toggles off, back, exit.
	l, _ := tui.NewTestLoop(dir, "en", "5a\rHi\r\r qq")
	if code := l.Loop(tui.NoopRestore); code != 0 {
		t.Fatalf(tui.ExitCodeFormat, code)
	}
	s, err := config.Load(dir)
	if err != nil {
		t.Fatalf(tui.LoadFailedFmt, err)
	}
	if len(s.CustomAlerts) != 1 {
		t.Fatalf("want 1 custom, got %+v", s.CustomAlerts)
	}
	c := s.CustomAlerts[0]
	if c.ID != "custom-1" || c.Event != "session.created" || c.Title != "Hi" || c.Enabled {
		t.Fatalf("unexpected custom state: %+v", c)
	}
}

func TestFullscreenCustomDelete(t *testing.T) {
	dir := t.TempDir()
	// Create, then D delete with y confirm, back, exit.
	l, _ := tui.NewTestLoop(dir, "en", "5a\rHi\r\rd"+"y\r"+"qq")
	if code := l.Loop(tui.NoopRestore); code != 0 {
		t.Fatalf(tui.ExitCodeFormat, code)
	}
	s, err := config.Load(dir)
	if err != nil {
		t.Fatalf(tui.LoadFailedFmt, err)
	}
	if len(s.CustomAlerts) != 0 {
		t.Fatalf("custom must be deleted, got %+v", s.CustomAlerts)
	}
}

func TestRunSmartFallsBackOffConsole(t *testing.T) {
	dir := t.TempDir()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe failed: %v", err)
	}
	defer r.Close()
	if _, err := w.WriteString("5\n"); err != nil {
		t.Fatalf("seed failed: %v", err)
	}
	w.Close()
	var out bytes.Buffer
	// Pipes are not consoles, so RunSmart must run the line mode.
	if code := tui.RunSmart(dir, r, &out, nil); code != 0 {
		t.Fatalf("fallback exit code = %d, want 0", code)
	}
	if !strings.Contains(out.String(), "Salir") && !strings.Contains(out.String(), "Exit") {
		t.Fatal("fallback must render the line-mode menu")
	}
}
