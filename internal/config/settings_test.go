package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultsAreV1Compatible(t *testing.T) {
	s := Defaults()
	if s.Lang != "en" {
		t.Fatalf("default lang = %q, want %q", s.Lang, "en")
	}
	if !s.Alerts.SessionIdle || !s.Alerts.SessionError || !s.Alerts.PermissionAsked || !s.Alerts.QuestionAsked {
		t.Fatalf("defaults must enable every alert, got %+v", s.Alerts)
	}
	if got := SettingsPath(`C:\temp\ocn\plugins`); got != filepath.Join(`C:\temp\ocn\plugins`, SettingsFileName) {
		t.Fatalf("settings path = %q", got)
	}
}

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	dir := t.TempDir()
	s, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s != Defaults() {
		t.Fatalf("got %+v, want defaults %+v", s, Defaults())
	}
}

func TestSaveRoundTrip(t *testing.T) {
	dir := t.TempDir()
	want := Defaults()
	want.Lang = "es"
	want.Alerts.SessionIdle = false
	if err := Save(dir, want); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	got, err := Load(dir)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	want = Normalize(want)
	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestLoadNormalizesUnknownLang(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(SettingsPath(dir), []byte(`{"version":1,"lang":"fr","alerts":{"sessionIdle":true,"sessionError":true,"permissionAsked":true,"questionAsked":true}}`), 0o644); err != nil {
		t.Fatalf("seed failed: %v", err)
	}
	s, err := Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Lang != "en" {
		t.Fatalf("lang = %q, want %q", s.Lang, "en")
	}
}

func TestLoadCorruptFileReturnsError(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(SettingsPath(dir), []byte(`{not json`), 0o644); err != nil {
		t.Fatalf("seed failed: %v", err)
	}
	if _, err := Load(dir); err == nil {
		t.Fatal("expected error, got nil")
	}
}
