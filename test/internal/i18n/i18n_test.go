package i18n_test

import (
	"testing"

	"github.com/IbatoLionDev/opencode-console-notify/internal/i18n"
)

func TestFallbackToEnglish(t *testing.T) {
	if got := i18n.T("fr", "menu.language"); got != "Language" {
		t.Fatalf("got %q, want English fallback", got)
	}
	if got := i18n.T("en", "missing.key"); got != "missing.key" {
		t.Fatalf("missing key should echo, got %q", got)
	}
}

func TestSpanishCoverage(t *testing.T) {
	for key := range i18n.En {
		if _, ok := i18n.Es[key]; !ok {
			t.Fatalf("spanish dictionary misses key %q", key)
		}
	}
	if got := i18n.T("es", "menu.language"); got != "Idioma" {
		t.Fatalf("got %q, want Idioma", got)
	}
	if !i18n.Supported("es") || !i18n.Supported("en") || i18n.Supported("fr") {
		t.Fatal("supported set must be exactly en+es")
	}
}
