package notifier

import (
	"testing"
)

func TestEscapeToastXML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "plain text untouched", input: "Task finished", want: "Task finished"},
		{name: "ampersand", input: "a & b", want: "a &amp; b"},
		{name: "angle brackets", input: "<toast>", want: "&lt;toast&gt;"},
		{name: "quotes", input: `"hi" 'yo'`, want: "&quot;hi&quot; &apos;yo&apos;"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escapeToastXML(tt.input); got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
