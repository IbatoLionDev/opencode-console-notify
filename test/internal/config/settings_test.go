package config_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/IbatoLionDev/opencode-console-notify/internal/config"
)

func TestDefaultsAreV1Behavior(t *testing.T) {
	s := config.Defaults()
	if s.Version != config.SettingsVersion {
		t.Fatalf("default version = %d, want %d", s.Version, config.SettingsVersion)
	}
	if s.Lang != "en" {
		t.Fatalf("default lang = %q, want %q", s.Lang, "en")
	}
	if !s.Alerts.SessionIdle || !s.Alerts.SessionError || !s.Alerts.PermissionAsked || !s.Alerts.QuestionAsked {
		t.Fatalf("defaults must enable every alert, got %+v", s.Alerts)
	}
	if len(s.CustomAlerts) != 0 {
		t.Fatalf("defaults must carry no customs, got %+v", s.CustomAlerts)
	}
	if got := config.SettingsPath(`C:\temp\ocn\plugins`); got != filepath.Join(`C:\temp\ocn\plugins`, config.SettingsFileName) {
		t.Fatalf("settings path = %q", got)
	}
}

func TestLoadMissingFileReturnsDefaults(t *testing.T) {
	dir := t.TempDir()
	s, err := config.Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !reflect.DeepEqual(s, config.Defaults()) {
		t.Fatalf("got %+v, want defaults %+v", s, config.Defaults())
	}
}

func TestSaveRoundTrip(t *testing.T) {
	dir := t.TempDir()
	want := config.Defaults()
	want.Lang = "es"
	want.Alerts.SessionIdle = false
	if err := config.Save(dir, want); err != nil {
		t.Fatalf("save failed: %v", err)
	}
	got, err := config.Load(dir)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	want = config.Normalize(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestLoadNormalizesUnknownLang(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(config.SettingsPath(dir), []byte(`{"version":1,"lang":"fr","alerts":{"sessionIdle":true,"sessionError":true,"permissionAsked":true,"questionAsked":true}}`), 0o644); err != nil {
		t.Fatalf("seed failed: %v", err)
	}
	s, err := config.Load(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Lang != "en" {
		t.Fatalf("lang = %q, want %q", s.Lang, "en")
	}
}

func TestLoadCorruptFileReturnsError(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(config.SettingsPath(dir), []byte(`{not json`), 0o644); err != nil {
		t.Fatalf("seed failed: %v", err)
	}
	if _, err := config.Load(dir); err == nil {
		t.Fatal("expected error, got nil")
	}
}
