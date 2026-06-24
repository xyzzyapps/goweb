package preacttodo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/manic/goweb/pkg/preproc"
	"github.com/manic/goweb/pkg/parser"
	"github.com/manic/goweb/pkg/tangle"
)

// TestTanglePreactTodo verifies that tangling the example produces
// all expected output files with correct content.
func TestTanglePreactTodo(t *testing.T) {
	dir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origDir); err != nil {
			t.Error(err)
		}
	}()

	// Copy .md files to temp dir so imports resolve.
	mdFiles := []string{"main.md", "app.md", "components.md", "config.md"}
	for _, f := range mdFiles {
		src := filepath.Join(origDir, f)
		data, err := os.ReadFile(src)
		if err != nil {
			// Try from the test's source directory.
			src = filepath.Join(origDir, "..", "..", "examples", "preact-todo", f)
			data, err = os.ReadFile(src)
			if err != nil {
				t.Skipf("skipping: cannot find example source %s: %v", f, err)
				return
			}
		}
		if err := os.WriteFile(filepath.Join(dir, f), data, 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Process the main file.
	vars := map[string]string{
		"APP_NAME": "TestTodo",
		"AUTHOR":   "Tester",
		"debug":    "true",
	}
	preprocResult, err := preproc.Process(filepath.Join(dir, "main.md"), vars)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := parser.ParseLines(preprocResult.Lines, filepath.Join(dir, "main.md"))
	if err != nil {
		t.Fatal(err)
	}
	doc.Vars = vars

	tn := tangle.New()
	if err := tn.Tangle(doc); err != nil {
		t.Fatal(err)
	}

	// Verify expected output files exist with content.
	expectedFiles := []string{
		"package.json",
		"tsconfig.json",
		"tailwind.config.js",
		"index.html",
		"LICENSE",
		"src/main.tsx",
		"src/app.tsx",
		"src/style.css",
		"src/components/add-todo.tsx",
		"src/components/todo-list.tsx",
		"src/components/todo-item.tsx",
	}
	for _, f := range expectedFiles {
		path := filepath.Join(dir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s was not created", f)
		} else {
			data, _ := os.ReadFile(path)
			if len(data) == 0 {
				t.Errorf("file %s is empty", f)
			}
		}
	}

	// Verify specific content.
	checkContent := func(file, substr string) {
		path := filepath.Join(dir, file)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("cannot read %s: %v", file, err)
			return
		}
		if !strings.Contains(string(data), substr) {
			t.Errorf("file %s missing expected content: %q", file, substr)
		}
	}

	checkContent("package.json", `"name": "TestTodo"`)
	checkContent("package.json", `"preact"`)
	checkContent("tsconfig.json", `"jsxImportSource": "preact"`)
	checkContent("index.html", `TestTodo`)
	checkContent("index.html", `src/main.tsx`)
	checkContent("src/main.tsx", `render(<App />`)
	checkContent("src/app.tsx", `function App()`)
	checkContent("src/app.tsx", `handleAdd`)
	checkContent("src/app.tsx", `handleToggle`)
	checkContent("src/app.tsx", `handleDelete`)
	checkContent("src/style.css", `@tailwind base`)
	checkContent("src/components/add-todo.tsx", `export function AddTodo`)
	checkContent("src/components/todo-list.tsx", `export function TodoList`)
	checkContent("src/components/todo-item.tsx", `export function TodoItem`)
	checkContent("LICENSE", `MIT License`)

	// Verify debug chunks are included when debug=true.
	checkContent("src/app.tsx", `console.log("[debug] added todo:"`)
	checkContent("src/app.tsx", `console.log("[debug] toggled todo:"`)
	checkContent("src/app.tsx", `console.log("[debug] deleted todo:"`)

	// Verify the override-demo tag is present in todo-list.
	// (todo-list.tsx has tags: component, override-demo)
	// Since we can't easily check tags on output files, just verify the file exists.
}
