package config_test

import (
	"strings"
	"testing"

	"github.com/IbatoLionDev/opencode-console-notify/internal/config"
)

// Shared fixtures so string literals are defined once.
const (
	testSessionDoneEvent = "session.execution.succeeded"
	testMissingEvent     = "nope.missing"
)

func TestCatalogHasDefaults(t *testing.T) {
	for _, want := range []string{testSessionDoneEvent, "session.execution.failed", "permission.asked", "form.created"} {
		if !config.IsKnownEvent(want) {
			t.Fatalf("catalog misses default event %q", want)
		}
	}
}

func TestCatalogHasNoDuplicates(t *testing.T) {
	seen := map[string]bool{}
	for _, e := range config.KnownEvents() {
		if e.Type == "" || e.DescEn == "" || e.DescEs == "" {
			t.Fatalf("catalog entry incomplete: %+v", e)
		}
		if seen[e.Type] {
			t.Fatalf("duplicate catalog event %q", e.Type)
		}
		seen[e.Type] = true
	}
	if len(seen) < 60 {
		t.Fatalf("catalog shrank to %d events, want at least 60", len(seen))
	}
}

func TestEventDescription(t *testing.T) {
	if got := config.EventDescription(testSessionDoneEvent, "es"); !strings.Contains(got, "esperando") {
		t.Fatalf("spanish description missing, got %q", got)
	}
	if got := config.EventDescription(testSessionDoneEvent, "fr"); got != config.EventDescription(testSessionDoneEvent, "en") {
		t.Fatalf("unknown lang must fall back to english, got %q", got)
	}
	if got := config.EventDescription(testMissingEvent, "en"); got != testMissingEvent {
		t.Fatalf("unknown event must echo, got %q", got)
	}
}

func TestNoisySubset(t *testing.T) {
	for _, want := range []string{"session.reasoning.delta", "session.tool.progress", "session.usage.updated", "skill.updated"} {
		if !config.IsNoisyEvent(want) {
			t.Fatalf("event %q must be flagged noisy", want)
		}
	}
	if config.IsNoisyEvent(testSessionDoneEvent) || config.IsNoisyEvent(testMissingEvent) {
		t.Fatal("quiet and unknown events must not be noisy")
	}
}
