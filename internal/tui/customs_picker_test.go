package tui

import (
	"strings"
	"testing"

	"github.com/IbatoLionDev/opencode-console-notify/internal/config"
	"github.com/IbatoLionDev/opencode-console-notify/internal/screen"
)

func firstCatalogMatch(t *testing.T, query string) int {
	t.Helper()
	q := strings.ToLower(query)
	for i, e := range config.KnownEvents() {
		if strings.Contains(strings.ToLower(e.Type), q) || strings.Contains(strings.ToLower(e.DescEn), q) {
			return i
		}
	}
	t.Fatalf("no event matches %q", query)
	return -1
}

func TestPickerRowsKeepsAllOnEmptyQuery(t *testing.T) {
	dir := t.TempDir()
	l, _ := newTestLoop(dir, "en", "")
	items, indexMap := l.pickerRows("")
	if len(items) != len(config.KnownEvents()) || len(indexMap) != len(items) {
		t.Fatalf("empty query must keep %d rows, got %d", len(config.KnownEvents()), len(items))
	}
	for i, idx := range indexMap {
		if idx != i {
			t.Fatalf("empty query must map identity, got %v", indexMap)
		}
	}
}

func TestPickerRowsMatchesAnywhereCaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	l, _ := newTestLoop(dir, "en", "")
	items, indexMap := l.pickerRows("SESS")
	if len(items) == 0 {
		t.Fatal("SESS must match session events case-insensitively")
	}
	for pos, item := range items {
		joined := strings.ToLower(item.Label + " " + item.Detail)
		if !strings.Contains(joined, "sess") {
			t.Fatalf("row %q does not contain sess", item.Label)
		}
		if indexMap[pos] < 0 || indexMap[pos] >= len(config.KnownEvents()) {
			t.Fatalf("map points outside the catalog: %v", indexMap)
		}
	}
	if _, indexMap := l.pickerRows("zzz-no-such-event"); len(indexMap) != 0 {
		t.Fatal("impossible query must match nothing")
	}
}

func TestPickerTypeKeyAppendsAndTrims(t *testing.T) {
	if nq, ok := pickerTypeKey(screen.ParsedKey{Key: screen.KeyRune, Rune: 's'}, ""); !ok || nq != "s" {
		t.Fatalf("rune must append, got %q %v", nq, ok)
	}
	// q arrives as quit from the parser but types inside the picker.
	if nq, ok := pickerTypeKey(screen.ParsedKey{Key: screen.KeyQuit, Rune: 'q'}, "s"); !ok || nq != "sq" {
		t.Fatalf("q must type, got %q %v", nq, ok)
	}
	if nq, ok := pickerTypeKey(screen.ParsedKey{Key: screen.KeyRune, Rune: 0x7f}, "sñ"); !ok || nq != "s" {
		t.Fatalf("backspace must trim rune-aware, got %q %v", nq, ok)
	}
	if _, ok := pickerTypeKey(screen.ParsedKey{Key: screen.KeyRune, Rune: 0x7f}, ""); ok {
		t.Fatal("backspace on empty query must not type")
	}
	if _, ok := pickerTypeKey(screen.ParsedKey{Key: screen.KeyUp}, "s"); ok {
		t.Fatal("arrows must not type")
	}
}

func TestPickEventFiltersThenPicks(t *testing.T) {
	dir := t.TempDir()
	l, _ := newTestLoop(dir, "en", "sess\r")
	want := firstCatalogMatch(t, "sess")
	if got := l.pickEvent(""); got != want {
		t.Fatalf("pickEvent(sess+Enter) = %d, want %d", got, want)
	}
}

func TestPickEventEscClearsThenBacksOut(t *testing.T) {
	dir := t.TempDir()
	l, _ := newTestLoop(dir, "en", "sess\x1b\x1b")
	if got := l.pickEvent(""); got != -1 {
		t.Fatalf("pickEvent(Esc twice) = %d, want -1", got)
	}
}

func TestPickEventEmptyMatchPicksNothing(t *testing.T) {
	dir := t.TempDir()
	l, _ := newTestLoop(dir, "en", "zzz-no-such-event\r")
	if got := l.pickEvent(""); got != -1 {
		t.Fatalf("pickEvent(no-match+Enter) = %d, want -1", got)
	}
}
