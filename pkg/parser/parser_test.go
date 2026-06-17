package parser

import (
	"testing"
)

func TestParseFenceLanguage(t *testing.T) {
	tests := []struct {
		line string
		want string
	}{
		{"```go", "go"},
		{"```python", "python"},
		{"```", ""},
		{"``` {chunk=\"main\"}", ""},
		{"~~~go", "go"},
	}
	for _, tt := range tests {
		got := parseFenceLanguage(tt.line)
		if got != tt.want {
			t.Errorf("parseFenceLanguage(%q) = %q; want %q", tt.line, got, tt.want)
		}
	}
}

func TestParseChunkHeader(t *testing.T) {
	tests := []struct {
		line     string
		wantName string
		wantAtt  map[string]string
	}{
		{
			line:     "<<main>>=",
			wantName: "main",
			wantAtt:  map[string]string{},
		},
		{
			line:     "<<main>>= file: main.go",
			wantName: "main",
			wantAtt:  map[string]string{"file": "main.go"},
		},
		{
			line:     "<<main>>= pipe: gofmt",
			wantName: "main",
			wantAtt:  map[string]string{"pipe": "gofmt"},
		},
		{
			line:     "<<main>>= file: main.go pipe: gofmt",
			wantName: "main",
			wantAtt:  map[string]string{"file": "main.go", "pipe": "gofmt"},
		},
		{
			line:     "<<main>>= override",
			wantName: "main",
			wantAtt:  map[string]string{"override": "true"},
		},
	}
	for _, tt := range tests {
		gotName, gotAtt := parseChunkHeader(tt.line)
		if gotName != tt.wantName {
			t.Errorf("parseChunkHeader(%q) name = %q; want %q", tt.line, gotName, tt.wantName)
		}
		if len(gotAtt) != len(tt.wantAtt) {
			t.Errorf("parseChunkHeader(%q) attrs = %v; want %v", tt.line, gotAtt, tt.wantAtt)
		}
		for k, v := range tt.wantAtt {
			if gotAtt[k] != v {
				t.Errorf("parseChunkHeader(%q) attr[%q] = %q; want %q", tt.line, k, gotAtt[k], v)
			}
		}
	}
}

func TestParseLines_Basic(t *testing.T) {
	lines := []string{
		"# Hello",
		"",
		"```go",
		"<<main>>=",
		`package main`,
		`func main() {}`,
		">>",
		"```",
	}
	doc, err := ParseLines(lines, "test.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(doc.Chunks))
	}
	if doc.Chunks[0].Name != "main" {
		t.Errorf("chunk name = %q; want %q", doc.Chunks[0].Name, "main")
	}
	if doc.Chunks[0].Language != "go" {
		t.Errorf("chunk language = %q; want %q", doc.Chunks[0].Language, "go")
	}
	if doc.Chunks[0].Body != "package main\nfunc main() {}" {
		t.Errorf("chunk body = %q; want %q", doc.Chunks[0].Body, "package main\nfunc main() {}")
	}
}

func TestParseLines_WithAttributes(t *testing.T) {
	lines := []string{
		"```go",
		"<<main>>= file: main.go pipe: gofmt",
		"package main",
		">>",
		"```",
	}
	doc, err := ParseLines(lines, "test.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(doc.Chunks))
	}
	if doc.Chunks[0].File != "main.go" {
		t.Errorf("chunk file = %q; want %q", doc.Chunks[0].File, "main.go")
	}
	if doc.Chunks[0].PipeCmd != "gofmt" {
		t.Errorf("chunk pipe = %q; want %q", doc.Chunks[0].PipeCmd, "gofmt")
	}
}

func TestParseLines_MultipleChunks(t *testing.T) {
	lines := []string{
		"```go",
		"<<a>>=",
		"chunk a",
		">>",
		"<<b>>=",
		"chunk b",
		">>",
		"```",
	}
	doc, err := ParseLines(lines, "test.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Chunks) != 2 {
		t.Fatalf("expected 2 chunks, got %d", len(doc.Chunks))
	}
	if doc.Chunks[0].Name != "a" {
		t.Errorf("first chunk name = %q; want %q", doc.Chunks[0].Name, "a")
	}
	if doc.Chunks[1].Name != "b" {
		t.Errorf("second chunk name = %q; want %q", doc.Chunks[1].Name, "b")
	}
}

func TestParseLines_References(t *testing.T) {
	chunk := &Chunk{
		Name: "main",
		Body: `package main
<<imports>>
func main() { <<body>> }`,
	}
	refs := extractReferences(chunk.Body)
	if len(refs) != 2 {
		t.Fatalf("expected 2 references, got %v", refs)
	}
	if refs[0] != "imports" {
		t.Errorf("first ref = %q; want %q", refs[0], "imports")
	}
	if refs[1] != "body" {
		t.Errorf("second ref = %q; want %q", refs[1], "body")
	}
}

func TestParseLines_NoCodeBlocks(t *testing.T) {
	lines := []string{
		"# Just documentation",
		"Nothing to see here.",
	}
	doc, err := ParseLines(lines, "test.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Chunks) != 0 {
		t.Errorf("expected 0 chunks, got %d", len(doc.Chunks))
	}
}

func TestParseLines_OutsideFence(t *testing.T) {
	// Chunks defined outside fenced code blocks (true noweb-style).
	lines := []string{
		"# Documentation",
		"",
		"<<main>>= file: main.go",
		"package main",
		"func main() {}",
		">>",
		"",
		"More docs.",
	}
	doc, err := ParseLines(lines, "test.md")
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(doc.Chunks))
	}
	if doc.Chunks[0].Name != "main" {
		t.Errorf("chunk name = %q; want %q", doc.Chunks[0].Name, "main")
	}
	if doc.Chunks[0].File != "main.go" {
		t.Errorf("chunk file = %q; want %q", doc.Chunks[0].File, "main.go")
	}
	if doc.Chunks[0].Body != "package main\nfunc main() {}" {
		t.Errorf("chunk body = %q", doc.Chunks[0].Body)
	}
}

func TestValidateDocument(t *testing.T) {
	doc := &Document{
		Chunks: []*Chunk{
			{Name: "a", Body: "<<b>>"},
		},
	}
	err := ValidateDocument(doc)
	if err == nil {
		t.Fatal("expected error for missing reference")
	}
}
