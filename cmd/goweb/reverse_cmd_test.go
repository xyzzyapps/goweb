package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCollectReverseFilesWalksDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "src"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "node_modules"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "src", "a.go"), []byte("package a\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "src", "b.go"), []byte("package b\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "node_modules", "x.js"), []byte("nope\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bin.dat"), []byte{0, 1, 2}, 0644); err != nil {
		t.Fatal(err)
	}

	files, err := collectReverseFiles([]string{dir}, "")
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(files, "\n")
	if !strings.Contains(joined, "a.go") || !strings.Contains(joined, "b.go") {
		t.Fatalf("missing sources: %v", files)
	}
	for _, f := range files {
		if strings.Contains(f, "node_modules") || strings.HasSuffix(f, "bin.dat") {
			t.Fatalf("should skip %s", f)
		}
	}
}

func TestUniqueChunkName(t *testing.T) {
	used := map[string]int{}
	a := uniqueChunkName("src/a.go", used)
	b := uniqueChunkName("lib/a.go", used)
	if a != "a" || b != "a-2" {
		t.Fatalf("got %q %q", a, b)
	}
}
