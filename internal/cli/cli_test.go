package cli

import (
	"bytes"
	"testing"

	"github.com/IbatoLionDev/opencode-console-notify/internal/aumid"
)

const pluginsDirFlag = "--plugins-dir"

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
