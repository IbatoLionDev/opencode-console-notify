package cli

import (
	"bytes"
	"strings"
	"testing"

	"github.com/IbatoLionDev/opencode-console-notify/internal/aumid"
	"github.com/IbatoLionDev/opencode-console-notify/internal/config"
	"github.com/IbatoLionDev/opencode-console-notify/internal/doctor"
)

const pluginsDirFlag = "--plugins-dir"

// configLangFlag is the language flag spelled the config tests use.
const configLangFlag = "--lang"

func checkParseArgsCase(t *testing.T, args []string, wantCmd, wantDir string, wantErr, wantUsage bool) {
	t.Helper()
	cmd, dir, err := parseArgs(args)
	if wantUsage {
		if err == nil {
			t.Fatal("expected usage signal, got nil error")
		}
		return
	}
	if wantErr {
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		return
	}
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd != wantCmd || dir != wantDir {
		t.Fatalf("got (%q, %q), want (%q, %q)", cmd, dir, wantCmd, wantDir)
	}
}

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantCmd   string
		wantDir   string
		wantErr   bool
		wantUsage bool // io.EOF: no command at all
	}{
		{name: "bare install", args: []string{"install"}, wantCmd: "install"},
		{name: "bare upgrade", args: []string{"upgrade"}, wantCmd: "upgrade"},
		{name: "upgrade flag before command", args: []string{pluginsDirFlag, `C:\p`, "upgrade"}, wantCmd: "upgrade", wantDir: `C:\p`},
		{name: "upgrade flag after command", args: []string{"upgrade", pluginsDirFlag, `C:\p`}, wantCmd: "upgrade", wantDir: `C:\p`},
		{name: "flag before command", args: []string{pluginsDirFlag, `C:\p`, "doctor"}, wantCmd: "doctor", wantDir: `C:\p`},
		{name: "flag after command", args: []string{"doctor", pluginsDirFlag, `C:\p`}, wantCmd: "doctor", wantDir: `C:\p`},
		{name: "equals form", args: []string{pluginsDirFlag + `=C:\p`, "test"}, wantCmd: "test", wantDir: `C:\p`},
		{name: "unknown command", args: []string{"frobnicate"}, wantErr: true},
		{name: "unknown flag", args: []string{"--frobnicate", "install"}, wantErr: true},
		{name: "flag without value", args: []string{"install", pluginsDirFlag}, wantErr: true},
		{name: "extra positional", args: []string{"install", "extra"}, wantErr: true},
		{name: "no args prints usage", args: nil, wantUsage: true},
		{name: "help", args: []string{"help"}, wantCmd: "help"},
		{name: "help long flag", args: []string{"--help"}, wantCmd: "help"},
		{name: "help short flag", args: []string{"-h"}, wantCmd: "help"},
		{name: "version", args: []string{"version"}, wantCmd: "version"},
		{name: "version long flag", args: []string{"--version"}, wantCmd: "version"},
		{name: "version short flag", args: []string{"-V"}, wantCmd: "version"},
		{name: "flag after command", args: []string{"install", "--version"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkParseArgsCase(t, tt.args, tt.wantCmd, tt.wantDir, tt.wantErr, tt.wantUsage)
		})
	}
}

func TestRunDoctorEndToEnd(t *testing.T) {
	// Hermetic backend probe: the machine PATH must never flip this test.
	old := doctor.NotifierLookPath
	doctor.NotifierLookPath = func(string) (string, error) { return "/bin/notify-send", nil }
	t.Cleanup(func() { doctor.NotifierLookPath = old })

	plugins := t.TempDir()
	reg := &aumid.FakeRegistry{}
	var out, errOut bytes.Buffer

	if code := Run([]string{"doctor", pluginsDirFlag, plugins}, reg, &out, &errOut); code == 0 {
		t.Fatal("run doctor before install must exit non-zero")
	}
	if code := Run([]string{"install", pluginsDirFlag, plugins}, reg, &out, &errOut); code != 0 {
		t.Fatalf("run install exited %d: %s", code, errOut.String())
	}
	out.Reset()
	if code := Run([]string{"doctor", pluginsDirFlag, plugins}, reg, &out, &errOut); code != 0 {
		t.Fatalf("run doctor after install exited %d:\n%s", code, out.String())
	}
	out.Reset()
	if code := Run([]string{"uninstall", pluginsDirFlag, plugins}, reg, &out, &errOut); code != 0 {
		t.Fatalf("run uninstall exited %d: %s", code, errOut.String())
	}
}

func TestParseConfigArgs(t *testing.T) {
	cmd, _, rest, err := parseArgsFull([]string{"config", configLangFlag, "es"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd != "config" || len(rest) != 2 || rest[0] != configLangFlag || rest[1] != "es" {
		t.Fatalf("got (%q, %q)", cmd, rest)
	}
	if _, _, _, err := parseArgsFull([]string{"install", "extra"}); err == nil {
		t.Fatal("expected error for extra positional on install")
	}
}

func TestRunConfigFlags(t *testing.T) {
	plugins := t.TempDir()
	reg := &aumid.FakeRegistry{}
	var out, errOut bytes.Buffer

	if code := Run([]string{"config", pluginsDirFlag, plugins, configLangFlag, "es"}, reg, &out, &errOut); code != 0 {
		t.Fatalf("config --lang exited %d: %s", code, errOut.String())
	}
	out.Reset()
	if code := Run([]string{"config", pluginsDirFlag, plugins, "--toggle", "sessionIdle=off"}, reg, &out, &errOut); code != 0 {
		t.Fatalf("config --toggle exited %d: %s", code, errOut.String())
	}
	out.Reset()
	if code := Run([]string{"config", pluginsDirFlag, plugins, "--list-alerts"}, reg, &out, &errOut); code != 0 {
		t.Fatalf("config --list-alerts exited %d: %s", code, errOut.String())
	}
	if code := Run([]string{"config", pluginsDirFlag, plugins, "--toggle", "bogus=on"}, reg, &out, &errOut); code == 0 {
		t.Fatal("config --toggle bogus must exit non-zero")
	}
}

func TestRunConfigListFlags(t *testing.T) {
	plugins := t.TempDir()
	reg := &aumid.FakeRegistry{}
	var out, errOut bytes.Buffer

	out.Reset()
	if code := Run([]string{"config", pluginsDirFlag, plugins, "--list-events"}, reg, &out, &errOut); code != 0 {
		t.Fatalf("config --list-events exited %d: %s", code, errOut.String())
	}
	for _, want := range []string{"session.idle", "file.edited", "todo.updated"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("--list-events misses %q", want)
		}
	}
	out.Reset()
	if code := Run([]string{"config", pluginsDirFlag, plugins, "--list-customs"}, reg, &out, &errOut); code != 0 {
		t.Fatalf("config --list-customs exited %d: %s", code, errOut.String())
	}
	s, err := config.Load(plugins)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if _, err := config.AddCustom(&s, "todo.updated", "Todos", ""); err != nil {
		t.Fatalf("seed failed: %v", err)
	}
	if err := config.Save(plugins, s); err != nil {
		t.Fatalf("seed save failed: %v", err)
	}
	out.Reset()
	if code := Run([]string{"config", pluginsDirFlag, plugins, "--list-customs"}, reg, &out, &errOut); code != 0 {
		t.Fatalf("config --list-customs exited %d: %s", code, errOut.String())
	}
	for _, want := range []string{"custom-1", "Todos", "[custom]"} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("--list-customs misses %q in:\n%s", want, out.String())
		}
	}
}
