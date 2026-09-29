package cli

import (
	"bytes"
	"testing"

	"github.com/IbatoLionDev/opencode-console-notify/internal/aumid"
)

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
		{name: "flag before command", args: []string{"--plugins-dir", `C:\p`, "doctor"}, wantCmd: "doctor", wantDir: `C:\p`},
		{name: "flag after command", args: []string{"doctor", "--plugins-dir", `C:\p`}, wantCmd: "doctor", wantDir: `C:\p`},
		{name: "equals form", args: []string{"--plugins-dir=C:\\p", "test"}, wantCmd: "test", wantDir: `C:\p`},
		{name: "unknown command", args: []string{"frobnicate"}, wantErr: true},
		{name: "unknown flag", args: []string{"--frobnicate", "install"}, wantErr: true},
		{name: "flag without value", args: []string{"install", "--plugins-dir"}, wantErr: true},
		{name: "extra positional", args: []string{"install", "extra"}, wantErr: true},
		{name: "no args prints usage", args: nil, wantUsage: true},
		{name: "help", args: []string{"help"}, wantCmd: "help"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, dir, err := parseArgs(tt.args)
			if tt.wantUsage {
				if err == nil {
					t.Fatal("expected usage signal, got nil error")
				}
				return
			}
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cmd != tt.wantCmd || dir != tt.wantDir {
				t.Fatalf("got (%q, %q), want (%q, %q)", cmd, dir, tt.wantCmd, tt.wantDir)
			}
		})
	}
}

func TestRunDoctorEndToEnd(t *testing.T) {
	plugins := t.TempDir()
	reg := &aumid.FakeRegistry{}
	var out, errOut bytes.Buffer

	if code := Run([]string{"doctor", "--plugins-dir", plugins}, reg, &out, &errOut); code == 0 {
		t.Fatal("run doctor before install must exit non-zero")
	}
	if code := Run([]string{"install", "--plugins-dir", plugins}, reg, &out, &errOut); code != 0 {
		t.Fatalf("run install exited %d: %s", code, errOut.String())
	}
	out.Reset()
	if code := Run([]string{"doctor", "--plugins-dir", plugins}, reg, &out, &errOut); code != 0 {
		t.Fatalf("run doctor after install exited %d:\n%s", code, out.String())
	}
	out.Reset()
	if code := Run([]string{"uninstall", "--plugins-dir", plugins}, reg, &out, &errOut); code != 0 {
		t.Fatalf("run uninstall exited %d: %s", code, errOut.String())
	}
}
