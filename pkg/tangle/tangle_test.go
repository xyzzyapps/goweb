package tangle

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/manic/goweb/pkg/parser"
)

func TestTangle(t *testing.T) {
	doc := &parser.Document{
		Chunks: []*parser.Chunk{
			{Name: "greeting", Body: `"Hello"`},
			{Name: "main", Body: "fmt.Println(<<greeting>>)", File: "output.go"},
		},
	}

	// Change to temp dir for file output.
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Error(err)
		}
	}()

	tangle := New()
	var buf bytes.Buffer
	tangle.Stdout = &buf

	if err := tangle.Tangle(doc); err != nil {
		t.Fatal(err)
	}

	// Check file was written.
	data, err := os.ReadFile(filepath.Join(dir, "output.go"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(data)) != `fmt.Println("Hello")` {
		t.Errorf("output.go = %q; want %q", string(data), `fmt.Println("Hello")`)
	}
}

func TestTangleChunk(t *testing.T) {
	doc := &parser.Document{
		Chunks: []*parser.Chunk{
			{Name: "a", Body: "hello"},
			{Name: "b", Body: "<<a>> world"},
		},
	}

	tangle := New()
	var buf bytes.Buffer
	if err := tangle.TangleChunk(doc, "b", &buf); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "hello world" {
		t.Errorf("got %q; want %q", buf.String(), "hello world")
	}
}

func TestSplitCommand(t *testing.T) {
	tests := []struct {
		cmd  string
		want []string
	}{
		{"gofmt", []string{"gofmt"}},
		{"gofmt -w", []string{"gofmt", "-w"}},
		{`"my program" --flag`, []string{"my program", "--flag"}},
	}
	for _, tt := range tests {
		got := splitCommand(tt.cmd)
		if len(got) != len(tt.want) {
			t.Errorf("splitCommand(%q) = %v; want %v", tt.cmd, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("splitCommand(%q) = %v; want %v", tt.cmd, got, tt.want)
			}
		}
	}
}

func TestFindChunk(t *testing.T) {
	doc := &parser.Document{
		Chunks: []*parser.Chunk{
			{Name: "a"},
			{Name: "b"},
		},
	}
	if c := findChunk(doc, "a"); c == nil || c.Name != "a" {
		t.Error("expected to find chunk 'a'")
	}
	if c := findChunk(doc, "nonexistent"); c != nil {
		t.Error("expected nil for nonexistent chunk")
	}
}
