// Settings persistence for the config command (v2.0.0).
// Missing file means current behavior: everything on, English.
// Schema v2 adds custom alerts; v1 files migrate silently.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SettingsFileName is the only settings file this CLI reads or writes.
// It lives next to the plugin file inside the plugins directory.
const SettingsFileName = "console-notify.config.json"

// SettingsVersion is the schema version written by Save.
const SettingsVersion = 2

// AlertSettings holds one on/off toggle per default notification.
type AlertSettings struct {
	SessionIdle     bool `json:"sessionIdle"`
	SessionError    bool `json:"sessionError"`
	PermissionAsked bool `json:"permissionAsked"`
	QuestionAsked   bool `json:"questionAsked"`
}

// Settings is the persisted config command state.
type Settings struct {
	Version      int           `json:"version"`
	Lang         string        `json:"lang"`
	Alerts       AlertSettings `json:"alerts"`
	CustomAlerts []CustomAlert `json:"customAlerts,omitempty"`
}

// Defaults returns the v1-compatible behavior: all alerts on, English.
func Defaults() Settings {
	return Settings{
		Version: SettingsVersion,
		Lang:    "en",
		Alerts: AlertSettings{
			SessionIdle:     true,
			SessionError:    true,
			PermissionAsked: true,
			QuestionAsked:   true,
		},
	}
}

// Normalize replaces unknown languages with English, pins the version,
// and drops custom entries that can never fire (unknown event or empty
// title) so one bad entry never breaks the file.
func Normalize(s Settings) Settings {
	if s.Lang != "en" && s.Lang != "es" {
		s.Lang = "en"
	}
	s.Version = SettingsVersion
	kept := s.CustomAlerts[:0]
	for _, c := range s.CustomAlerts {
		if ValidateCustom(c.Event, c.Title) == nil {
			kept = append(kept, c)
		}
	}
	s.CustomAlerts = kept
	return s
}

// SettingsPath joins the plugins directory with the settings file name.
func SettingsPath(pluginsDir string) string {
	return filepath.Join(pluginsDir, SettingsFileName)
}

// Load reads the settings file. A missing file returns Defaults with no
// error so fresh installs keep the current behavior. A corrupt file
// returns an error in English and leaves the install untouched.
func Load(pluginsDir string) (Settings, error) {
	path := SettingsPath(pluginsDir)
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Defaults(), nil
		}
		return Settings{}, fmt.Errorf("read settings file: %w", err)
	}
	var s Settings
	if err := json.Unmarshal(raw, &s); err != nil {
		return Settings{}, fmt.Errorf("parse settings file: %w", err)
	}
	return Normalize(s), nil
}

// Save writes settings atomically via a temp file in the same directory
// followed by a rename, so a running OpenCode never observes a half-written
// file. Any leftover temp file is cleaned up.
func Save(pluginsDir string, s Settings) error {
	s = Normalize(s)
	if err := os.MkdirAll(pluginsDir, 0o755); err != nil {
		return fmt.Errorf("create plugins directory: %w", err)
	}
	raw, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("encode settings file: %w", err)
	}
	raw = append(raw, '\n')
	dest := SettingsPath(pluginsDir)
	tmp, err := os.CreateTemp(filepath.Dir(dest), SettingsFileName+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp file: %w", err)
	}
	if err := tmp.Chmod(0o644); err != nil {
		tmp.Close()
		return fmt.Errorf("chmod temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Rename(tmpName, dest); err != nil {
		return fmt.Errorf("rename temp file into place: %w", err)
	}
	return nil
}
