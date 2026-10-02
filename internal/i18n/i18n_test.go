package i18n

import "testing"

func TestFallbackToEnglish(t *testing.T) {
	if got := T("fr", "menu.language"); got != "Language" {
		t.Fatalf("got %q, want English fallback", got)
	}
	if got := T("en", "missing.key"); got != "missing.key" {
		t.Fatalf("missing key should echo, got %q", got)
	}
}

func TestSpanishCoverage(t *testing.T) {
	for key := range en {
		if _, ok := es[key]; !ok {
			t.Fatalf("spanish dictionary misses key %q", key)
		}
	}
	if got := T("es", "menu.language"); got != "Idioma" {
		t.Fatalf("got %q, want Idioma", got)
	}
	if !Supported("es") || !Supported("en") || Supported("fr") {
		t.Fatal("supported set must be exactly en+es")
	}
}
