package weave

// Tests are currently focused on stripControlSyntax.
// The Weave function signature is tested via integration tests.

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
			name:  "chunk def wrapped in fences",
			lines: []string{"<<main>>=", "body", ">>"},
			want:  "```\nbody\n```",
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
			name: "chunk def with file: attribute sets language",
			lines: []string{"<<main>>= file: main.go", "package main", ">>"},
			want:  "```go\npackage main\n```",
		},
		{
			name: "chunk def with tsx file attribute",
			lines: []string{"<<comp>>= file: src/app.tsx tags: component", "import { h } from 'preact'", ">>"},
			want:  "```tsx\nimport { h } from 'preact'\n```",
		},
		{
			name: "nested defs inside fenced block preserved",
			lines: []string{"```", "<<outer>>=", "outer body", ">>", "```"},
			want:  "```\nouter body\n```",
		},
		{
			name: "fenced block with language preserved",
			lines: []string{"```go", "func main() {}", "```"},
			want:  "```go\nfunc main() {}\n```",
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
