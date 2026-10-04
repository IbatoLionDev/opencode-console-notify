package config

import (
	"os"
	"testing"
)

// Shared event fixtures so string literals are defined once.
const (
	testFileEditedEvent  = "file.edited"
	testTodoUpdatedEvent = "todo.updated"
)

func TestMigrateV1KeepsBehavior(t *testing.T) {
	dir := t.TempDir()
	raw := `{"version":1,"lang":"es","alerts":{"sessionIdle":false,"sessionError":true,"permissionAsked":true,"questionAsked":true}}`
	if err := os.WriteFile(SettingsPath(dir), []byte(raw), 0o644); err != nil {
		t.Fatalf("seed failed: %v", err)
	}
	s, err := Load(dir)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if s.Version != SettingsVersion || s.Lang != "es" || s.Alerts.SessionIdle {
		t.Fatalf("migration broke v1 values: %+v", s)
	}
	if len(s.CustomAlerts) != 0 {
		t.Fatalf("v1 file must load with no customs, got %+v", s.CustomAlerts)
	}
}

func TestCustomCRUD(t *testing.T) {
	s := Defaults()
	id, err := AddCustom(&s, testFileEditedEvent, " Edited ", "saved")
	if err != nil {
		t.Fatalf("add failed: %v", err)
	}
	if id != "custom-1" || s.CustomAlerts[0].Title != "Edited" {
		t.Fatalf("unexpected add result: %+v", s.CustomAlerts)
	}
	if !s.CustomAlerts[0].Enabled {
		t.Fatal("new customs must start enabled")
	}
	if ok, err := UpdateCustom(&s, id, testTodoUpdatedEvent, "Todos", ""); !ok || err != nil {
		t.Fatalf("update failed: ok=%v err=%v", ok, err)
	}
	if !ToggleCustom(&s, id) || FindCustom(s, id).Enabled {
		t.Fatal("toggle must disable")
	}
	if !RemoveCustom(&s, id) || FindCustom(s, id) != nil {
		t.Fatal("remove must delete")
	}
	if RemoveCustom(&s, id) || ToggleCustom(&s, id) {
		t.Fatal("missing ids must report false")
	}
	if ok, _ := UpdateCustom(&s, id, testTodoUpdatedEvent, "x", ""); ok {
		t.Fatal("update of missing id must report false")
	}
}

func TestCustomValidation(t *testing.T) {
	s := Defaults()
	if _, err := AddCustom(&s, "nope.missing", "T", ""); err == nil {
		t.Fatal("unknown event must be rejected")
	}
	if _, err := AddCustom(&s, testFileEditedEvent, "   ", ""); err == nil {
		t.Fatal("empty title must be rejected")
	}
	if len(s.CustomAlerts) != 0 {
		t.Fatal("rejected adds must not mutate")
	}
}

func TestInvalidCustomsFilteredOnLoad(t *testing.T) {
	dir := t.TempDir()
	raw := `{"version":2,"lang":"en","alerts":{"sessionIdle":true,"sessionError":true,"permissionAsked":true,"questionAsked":true},"customAlerts":[{"id":"custom-1","event":"nope.missing","title":"T","body":"","enabled":true},{"id":"custom-2","event":"file.edited","title":"F","body":"","enabled":true}]}`
	if err := os.WriteFile(SettingsPath(dir), []byte(raw), 0o644); err != nil {
		t.Fatalf("seed failed: %v", err)
	}
	s, err := Load(dir)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(s.CustomAlerts) != 1 || s.CustomAlerts[0].ID != "custom-2" {
		t.Fatalf("invalid entries must be filtered, got %+v", s.CustomAlerts)
	}
}

func TestCustomsForEvent(t *testing.T) {
	s := Defaults()
	if _, err := AddCustom(&s, testFileEditedEvent, "A", ""); err != nil {
		t.Fatal(err)
	}
	id2, err := AddCustom(&s, testFileEditedEvent, "B", "")
	if err != nil {
		t.Fatal(err)
	}
	ToggleCustom(&s, id2)
	got := CustomsForEvent(s, testFileEditedEvent)
	if len(got) != 1 || got[0].Title != "A" {
		t.Fatalf("want only enabled customs in order, got %+v", got)
	}
	if len(CustomsForEvent(s, testTodoUpdatedEvent)) != 0 {
		t.Fatal("other events must not match")
	}
}
