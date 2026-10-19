// Package tui testexport exposes identifiers needed by black-box tests
// living in test/internal/tui. Production code must not depend on this
// file; it exists only so the test package can probe internals without
// being in package tui.

package tui

import (
	"bytes"
	"strings"

	"github.com/IbatoLionDev/opencode-console-notify/internal/config"
	"github.com/IbatoLionDev/opencode-console-notify/internal/screen"
)

// ExitCodeFormat is the failure format shared by fullscreen tests.
const ExitCodeFormat = "exit code = %d, want 0"

// LoadFailedFmt is the failure format for config load assertions.
const LoadFailedFmt = "load failed: %v"

// MenuTarget maps a menu selection index to its view name.
func MenuTarget(sel int) string {
	return menuTarget(sel)
}

// NewTestLoop builds a scripted fullscreen loop for black-box tests.
func NewTestLoop(dir, lang, input string) (*ScreenLoop, *bytes.Buffer) {
	var out bytes.Buffer
	l := &ScreenLoop{
		PluginsDir: dir,
		Settings:   config.Defaults(),
		Lang:       lang,
		In:         strings.NewReader(input),
		Out:        &out,
		Width:      40,
		Height:     20,
	}
	l.Settings.Lang = lang
	return l, &out
}

// NoopRestore stands in for the console restore func: scripted loops never
// touch the real console, so there is nothing to restore.
func NoopRestore() {
	// Empty on purpose (see above): no console was enabled in tests.
}

// AssertScreenLoop exposes the loop's settings for black-box assertions.
func (l *ScreenLoop) AssertSettings() config.Settings {
	return l.Settings
}

// PickerTypeKey exposes pickerTypeKey for black-box tests.
func PickerTypeKey(k screen.ParsedKey, query string) (string, bool) {
	return pickerTypeKey(k, query)
}
