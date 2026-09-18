package fill

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFillMockAndSeal(t *testing.T) {
	dir := t.TempDir()
	md := filepath.Join(dir, "p.md")
	src := "# x\n\n<<add>>= file: add.js ai:open\nfunction add() {\n}\n>>\n\n<<keep>>= file: keep.js\nconst x = 1\n>>\n"
	if err := os.WriteFile(md, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	mock := filepath.Join(dir, "mock")
	if err := os.Mkdir(mock, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mock, "add.txt"), []byte("function add() { return 1 }\n"), 0644); err != nil {
		t.Fatal(err)
	}

	n, err := Fill(md, nil, Options{Mock: true, MockDir: mock})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("filled %d, want 1", n)
	}
	got, _ := os.ReadFile(md)
	s := string(got)
	if !strings.Contains(s, "ai:filled") {
		t.Fatalf("expected ai:filled:\n%s", s)
	}
	if !strings.Contains(s, "function add() { return 1 }") {
		t.Fatalf("expected mock body:\n%s", s)
	}
	if !strings.Contains(s, "const x = 1") {
		t.Fatalf("unmarked chunk rewritten:\n%s", s)
	}

	n, err = Seal(md, nil, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("sealed %d, want 1", n)
	}
	got, _ = os.ReadFile(md)
	if !strings.Contains(string(got), "ai:sealed") {
		t.Fatalf("expected ai:sealed:\n%s", got)
	}
}

func TestFillDryRun(t *testing.T) {
	dir := t.TempDir()
	md := filepath.Join(dir, "p.md")
	src := "<<h>>= ai:open\nstub\n>>\n"
	if err := os.WriteFile(md, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	n, err := Fill(md, nil, Options{DryRun: true, Mock: true})
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("dry-run count %d", n)
	}
	got, _ := os.ReadFile(md)
	if !strings.Contains(string(got), "ai:open") {
		t.Fatal("dry-run wrote the file")
	}
}

func TestSealOpenRequiresEmptyOK(t *testing.T) {
	dir := t.TempDir()
	md := filepath.Join(dir, "p.md")
	if err := os.WriteFile(md, []byte("<<h>>= ai:open\nstub\n>>\n"), 0644); err != nil {
		t.Fatal(err)
	}
	_, err := Seal(md, nil, Options{})
	if err == nil {
		t.Fatal("expected error")
	}
}
