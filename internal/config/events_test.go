package config

import (
	"strings"
	"testing"
)

// Shared fixtures so string literals are defined once.
const (
	testSessionIdleEvent = "session.idle"
	testMissingEvent     = "nope.missing"
)

func TestCatalogHasDefaults(t *testing.T) {
	for _, want := range []string{testSessionIdleEvent, "session.error", "permission.asked", "question.asked"} {
		if !IsKnownEvent(want) {
			t.Fatalf("catalog misses default event %q", want)
		}
	}
}

func TestCatalogHasNoDuplicates(t *testing.T) {
	seen := map[string]bool{}
	for _, e := range KnownEvents() {
		if e.Type == "" || e.DescEn == "" || e.DescEs == "" {
			t.Fatalf("catalog entry incomplete: %+v", e)
		}
		if seen[e.Type] {
			t.Fatalf("duplicate catalog event %q", e.Type)
		}
		seen[e.Type] = true
	}
	if len(seen) < 40 {
		t.Fatalf("catalog shrank to %d events, want at least 40", len(seen))
	}
}

func TestEventDescription(t *testing.T) {
	if got := EventDescription(testSessionIdleEvent, "es"); !strings.Contains(got, "esperando") {
		t.Fatalf("spanish description missing, got %q", got)
	}
	if got := EventDescription(testSessionIdleEvent, "fr"); got != EventDescription(testSessionIdleEvent, "en") {
		t.Fatalf("unknown lang must fall back to english, got %q", got)
	}
	if got := EventDescription(testMissingEvent, "en"); got != testMissingEvent {
		t.Fatalf("unknown event must echo, got %q", got)
	}
}

func TestNoisySubset(t *testing.T) {
	for _, want := range []string{"session.status", "message.part.updated", "file.watcher.updated", "tui.prompt.append"} {
		if !IsNoisyEvent(want) {
			t.Fatalf("event %q must be flagged noisy", want)
		}
	}
	if IsNoisyEvent(testSessionIdleEvent) || IsNoisyEvent(testMissingEvent) {
		t.Fatal("quiet and unknown events must not be noisy")
	}
}
