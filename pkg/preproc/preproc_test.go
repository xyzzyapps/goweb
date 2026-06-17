package preproc

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEvaluateCondition(t *testing.T) {
	vars := map[string]string{
		"debug": "true",
		"lang":  "go",
		"os":    "linux",
	}

	tests := []struct {
		expr string
		want bool
	}{
		{"debug", true}, // "true" is truthy
		{"lang", true},  // "go" is truthy
		{"nonexistent", false},
		{"debug==true", true},
		{"debug==false", false},
		{"lang==go", true},
		{"lang==python", false},
		{"lang!=python", true},
		{"lang!=go", false},
		{"nonexistent==", true},
		{"nonexistent!=value", true},
	}

	for _, tt := range tests {
		got := evaluateCondition(tt.expr, vars)
		if got != tt.want {
			t.Errorf("evaluateCondition(%q, vars) = %v; want %v", tt.expr, got, tt.want)
		}
	}
}

func TestParseConditionalDirective(t *testing.T) {
	tests := []struct {
		line     string
		wantType string
		wantExpr string
	}{
		{"<<if debug>>", "if", "debug"},
		{"<<elif lang==go>>", "elif", "lang==go"},
		{"<<else>>", "else", ""},
		{"<<end>>", "end", ""},
	}
	for _, tt := range tests {
		dirType, expr := parseConditionalDirective(tt.line)
		if dirType != tt.wantType || expr != tt.wantExpr {
			t.Errorf("parseConditionalDirective(%q) = (%q, %q); want (%q, %q)",
				tt.line, dirType, expr, tt.wantType, tt.wantExpr)
		}
	}
}

func TestProcess_Simple(t *testing.T) {
	// Create a temp file.
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	content := `# Hello
Some text
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Process(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %v", len(result.Lines), result.Lines)
	}
}

func TestProcess_ConditionalTrue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	content := `before
<<if debug>>
debug line
<<end>>
after
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Process(path, map[string]string{"debug": "true"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %v", len(result.Lines), result.Lines)
	}
	if result.Lines[1] != "debug line" {
		t.Errorf("expected 'debug line', got %q", result.Lines[1])
	}
}

func TestProcess_ConditionalFalse(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	content := `before
<<if debug==true>>
debug line
<<end>>
after
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Process(path, map[string]string{"debug": "false"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %v", len(result.Lines), result.Lines)
	}
}

func TestProcess_ConditionalNameCheck(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	content := `before
<<if debug>>
debug line
<<end>>
after
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Var exists and non-empty -> true
	result, err := Process(path, map[string]string{"debug": "true"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Lines) != 3 {
		t.Fatalf("expected 3 lines when debug=true, got %d: %v", len(result.Lines), result.Lines)
	}

	// Var doesn't exist -> false
	result, err = Process(path, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Lines) != 2 {
		t.Fatalf("expected 2 lines when debug unset, got %d: %v", len(result.Lines), result.Lines)
	}
}

func TestProcess_IfElse(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	content := `<<if lang=="go">>
go version
<<else>>
other version
<<end>>
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Process(path, map[string]string{"lang": "go"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Lines) != 1 || result.Lines[0] != "go version" {
		t.Fatalf("expected 'go version', got %v", result.Lines)
	}

	result, err = Process(path, map[string]string{"lang": "python"})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Lines) != 1 || result.Lines[0] != "other version" {
		t.Fatalf("expected 'other version', got %v", result.Lines)
	}
}

func TestProcess_Import(t *testing.T) {
	dir := t.TempDir()

	// Create imported file.
	importPath := filepath.Join(dir, "types.md")
	importContent := `<<types>>=
type MyType struct{}
>>`
	if err := os.WriteFile(importPath, []byte(importContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Create main file that imports.
	mainPath := filepath.Join(dir, "main.md")
	mainContent := `<<import "types.md">>
<<main>>=
func main() {}
>>`
	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatal(err)
	}

	result, err := Process(mainPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	// 3 lines from types.md + 3 lines from main.md = 6
	if len(result.Lines) != 6 {
		t.Fatalf("expected 6 lines, got %d: %v", len(result.Lines), result.Lines)
	}
}
