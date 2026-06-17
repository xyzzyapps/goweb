package graph

import (
	"testing"

	"github.com/manic/goweb/pkg/parser"
)

func TestFindReferences(t *testing.T) {
	tests := []struct {
		body string
		want map[string]bool
	}{
		{"no refs", map[string]bool{}},
		{"<<a>>", map[string]bool{"a": true}},
		{"<<a>> <<b>>", map[string]bool{"a": true, "b": true}},
		{"<<a>>= definition, not ref", map[string]bool{}},
		{"<<a>> <<b>>= <<c>>", map[string]bool{"a": true, "c": true}},
	}
	for _, tt := range tests {
		got := findReferences(tt.body)
		if len(got) != len(tt.want) {
			t.Errorf("findReferences(%q) = %v; want %v", tt.body, got, tt.want)
		}
		for k := range tt.want {
			if !got[k] {
				t.Errorf("findReferences(%q) missing key %q", tt.body, k)
			}
		}
	}
}

func TestTopologicalSort_Simple(t *testing.T) {
	doc := &parser.Document{
		Chunks: []*parser.Chunk{
			{Name: "a", Body: "just code"},
			{Name: "b", Body: "<<a>>"},
			{Name: "c", Body: "<<b>>"},
		},
	}
	g, err := New(doc)
	if err != nil {
		t.Fatal(err)
	}
	sorted, err := g.TopologicalSort()
	if err != nil {
		t.Fatal(err)
	}
	// a before b before c
	positions := make(map[string]int)
	for i, c := range sorted {
		positions[c.Name] = i
	}
	if positions["a"] > positions["b"] {
		t.Error("expected a before b")
	}
	if positions["b"] > positions["c"] {
		t.Error("expected b before c")
	}
}

func TestTopologicalSort_NoDeps(t *testing.T) {
	doc := &parser.Document{
		Chunks: []*parser.Chunk{
			{Name: "a", Body: "no deps"},
			{Name: "b", Body: "no deps"},
		},
	}
	g, err := New(doc)
	if err != nil {
		t.Fatal(err)
	}
	sorted, err := g.TopologicalSort()
	if err != nil {
		t.Fatal(err)
	}
	if len(sorted) != 2 {
		t.Fatalf("expected 2 sorted chunks, got %d", len(sorted))
	}
}

func TestTopologicalSort_Cycle(t *testing.T) {
	doc := &parser.Document{
		Chunks: []*parser.Chunk{
			{Name: "a", Body: "<<b>>"},
			{Name: "b", Body: "<<a>>"},
		},
	}
	g, err := New(doc)
	if err != nil {
		t.Fatal(err)
	}
	_, err = g.TopologicalSort()
	if err == nil {
		t.Fatal("expected cycle error")
	}
}

func TestResolve(t *testing.T) {
	doc := &parser.Document{
		Chunks: []*parser.Chunk{
			{Name: "greeting", Body: `"Hello"`},
			{Name: "main", Body: "fmt.Println(<<greeting>>)"},
		},
	}
	g, err := New(doc)
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := g.Resolve("main", nil)
	if err != nil {
		t.Fatal(err)
	}
	if resolved != `fmt.Println("Hello")` {
		t.Errorf("resolved = %q; want %q", resolved, `fmt.Println("Hello")`)
	}
}

func TestResolve_Circular(t *testing.T) {
	doc := &parser.Document{
		Chunks: []*parser.Chunk{
			{Name: "a", Body: "<<b>>"},
			{Name: "b", Body: "<<a>>"},
		},
	}
	g, err := New(doc)
	if err != nil {
		t.Fatal(err)
	}
	_, err = g.Resolve("a", nil)
	if err == nil {
		t.Fatal("expected circular reference error")
	}
}

func TestResolve_Missing(t *testing.T) {
	doc := &parser.Document{
		Chunks: []*parser.Chunk{
			{Name: "a", Body: "<<missing>>"},
		},
	}
	g, err := New(doc)
	if err != nil {
		t.Fatal(err)
	}
	_, err = g.Resolve("a", nil)
	if err == nil {
		t.Fatal("expected missing chunk error")
	}
}
