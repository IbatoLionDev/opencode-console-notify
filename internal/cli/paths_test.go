package cli

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestResolvePluginsDir(t *testing.T) {
	def, err := DefaultPluginsDir()
	if err != nil {
		t.Skipf("no home directory in this environment: %v", err)
	}

	tests := []struct {
		name     string
		override string
		want     string
		wantErr  bool
	}{
		{name: "empty override falls back to default", override: "", want: def},
		{name: "blank override falls back to default", override: "   ", want: def},
		{name: "explicit dir wins", override: `C:\temp\ocn\plugins`, want: `C:\temp\ocn\plugins`},
		{name: "explicit dir is cleaned", override: `C:\temp\ocn\sub\..\plugins`, want: `C:\temp\ocn\plugins`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolvePluginsDir(tt.override)
			if tt.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDefaultPluginsDirLayout(t *testing.T) {
	def, err := DefaultPluginsDir()
	if err != nil {
		t.Skipf("no home directory in this environment: %v", err)
	}
	wantSuffix := filepath.Join(".config", "opencode", "plugins")
	if !strings.HasSuffix(def, wantSuffix) {
		t.Fatalf("default %q does not end with %q", def, wantSuffix)
	}
	if got := PluginPath(def); !strings.HasSuffix(got, PluginFileName) {
		t.Fatalf("plugin path %q does not end with %q", got, PluginFileName)
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
