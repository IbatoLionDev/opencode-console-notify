package config

import (
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestResolvePluginsDir(t *testing.T) {
	def, err := DefaultPluginsDir()
	if err != nil {
		t.Skipf("no home directory in this environment: %v", err)
	}

	tests := []struct {
		name        string
		override    string
		want        string
		wantErr     bool
		windowsOnly bool // backslash paths only clean on Windows
	}{
		{name: "empty override falls back to default", override: "", want: def},
		{name: "blank override falls back to default", override: "   ", want: def},
		{name: "explicit dir wins", override: `C:\temp\ocn\plugins`, want: `C:\temp\ocn\plugins`, windowsOnly: true},
		{name: "explicit dir is cleaned", override: `C:\temp\ocn\sub\..\plugins`, want: `C:\temp\ocn\plugins`, windowsOnly: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.windowsOnly && runtime.GOOS != "windows" {
				t.Skip("backslash separators only exist on Windows")
			}
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
