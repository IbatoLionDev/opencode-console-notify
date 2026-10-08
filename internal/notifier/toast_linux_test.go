//go:build linux

package notifier

import (
	"strings"
	"testing"
)

func TestBuildNotifySendArgs(t *testing.T) {
	args := buildNotifySendArgs("OpenCode", "hello")
	if len(args) != 5 {
		t.Fatalf("want 5 argv entries, got %q", args)
	}
	if args[0] != "--app-name=OpenCode" {
		t.Fatalf("first arg must pin the app identity, got %q", args[0])
	}
	if args[3] != "OpenCode" || args[4] != "hello" {
		t.Fatalf("summary and body must travel verbatim, got %q", args)
	}
	if strings.Join(args, " ") == "" {
		t.Fatal("args must not render empty")
	}
}
