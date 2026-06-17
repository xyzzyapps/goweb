package parser

import (
	"testing"
)

func TestExtractReferences(t *testing.T) {
	tests := []struct {
		body string
		want []string
	}{
		{body: "hello <<name>> world", want: []string{"name"}},
		{body: "<<a>> <<b>> <<c>>", want: []string{"a", "b", "c"}},
		{body: "<<def>>=", want: nil},
		{body: "no refs here", want: nil},
		// Nested << inside a name is an edge case; the parser takes
		// the outermost << and the next >>, so <<a<<b>> becomes "a<<b".
		// This is fine for real usage where chunk names are simple identifiers.
	}
	for _, tt := range tests {
		got := extractReferences(tt.body)
		if !stringSliceEqual(got, tt.want) {
			t.Errorf("extractReferences(%q) = %v; want %v", tt.body, got, tt.want)
		}
	}
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestChunkTable_Duplicates(t *testing.T) {
	doc := &Document{
		Chunks: []*Chunk{
			{Name: "a", Source: "f1.md", Line: 1},
			{Name: "a", Source: "f2.md", Line: 5},
		},
	}
	_, err := doc.ChunkTable()
	if err == nil {
		t.Fatal("expected duplicate error")
	}
}

func TestMerge_Concatenate(t *testing.T) {
	a := &Chunk{Name: "x", Body: "part1"}
	b := &Chunk{Name: "x", Body: "part2"}
	a.Merge(b)
	if a.Body != "part1\npart2" {
		t.Errorf("expected 'part1\\npart2', got %q", a.Body)
	}
}

func TestMerge_Override(t *testing.T) {
	a := &Chunk{Name: "x", Body: "original"}
	b := &Chunk{Name: "x", Body: "replacement", Override: true}
	a.Merge(b)
	if a.Body != "replacement" {
		t.Errorf("expected 'replacement', got %q", a.Body)
	}
}

func TestChunksByFile(t *testing.T) {
	doc := &Document{
		Chunks: []*Chunk{
			{Name: "a", File: "out.go"},
			{Name: "b", File: ""},
			{Name: "c", File: "out.go"},
		},
	}
	byFile := doc.ChunksByFile()
	if len(byFile["out.go"]) != 2 {
		t.Errorf("expected 2 chunks for out.go, got %d", len(byFile["out.go"]))
	}
	if len(byFile[""]) != 1 {
		t.Errorf("expected 1 chunk for '', got %d", len(byFile[""]))
	}
}
