package weave

import (
	"strings"
	"testing"
)

func TestStripControlSyntax(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  string
	}{
		{
			name:  "simple text passes through",
			lines: []string{"hello", "world"},
			want:  "hello\nworld",
		},
		{
			name:  "chunk def line removed",
			lines: []string{"<<main>>=", "body", ">>"},
			want:  "body",
		},
		{
			name:  "references unwrapped",
			lines: []string{"<<ref>>"},
			want:  "ref",
		},
		{
			name:  "directives removed",
			lines: []string{"<<if debug>>", "content", "<<end>>"},
			want:  "content",
		},
		{
			name: "chunk def with attributes",
			// <<main>>= file: main.go  →  isChunkDefStart skips it entirely
			// (removeAngleBrackets would keep " file: main.go" but
			// stripeControlSyntax now catches it with isChunkDefStart)
			lines: []string{"<<main>>= file: main.go", "code", ">>"},
			want:  "code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := strings.Join(stripControlSyntax(tt.lines), "\n")
			if got != tt.want {
				t.Errorf("stripControlSyntax = %q; want %q", got, tt.want)
			}
		})
	}
}

func TestRemoveAngleBrackets(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{"plain text", "plain text"},
		{"<<ref>>", "ref"},
		{"<<main>>=", ""},
		{"fmt.Println(<<msg>>)", "fmt.Println(msg)"},
	}
	for _, tt := range tests {
		got := removeAngleBrackets(tt.line)
		if got != tt.want {
			t.Errorf("removeAngleBrackets(%q) = %q; want %q", tt.line, got, tt.want)
		}
	}
}

func TestIsChunkDefStart(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		{"<<main>>=", true},
		{"<<main>>= file: main.go", true},
		{"<<ref>>", false},
		{"plain", false},
	}
	for _, tt := range tests {
		got := isChunkDefStart(tt.line)
		if got != tt.want {
			t.Errorf("isChunkDefStart(%q) = %v; want %v", tt.line, got, tt.want)
		}
	}
}
