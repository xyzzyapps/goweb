package preacttodo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xyzzyapps/goweb/pkg/preproc"
	"github.com/xyzzyapps/goweb/pkg/parser"
	"github.com/xyzzyapps/goweb/pkg/tangle"
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
		"EMAIL":    "tester@example.com",
		"REPO":     "https://github.com/xyzzyapps/goweb",
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
		"index.html",
		"LICENSE",
		"src/main.js",
		"src/app.js",
		"src/style.css",
		"src/components/add-todo.js",
		"src/components/todo-list.js",
		"src/components/todo-item.js",
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
	checkContent("index.html", `TestTodo`)
	checkContent("index.html", `src/main.js`)
	checkContent("index.html", `importmap`)
	checkContent("src/main.js", `import { render } from "preact"`)
	checkContent("src/app.js", `function App()`)
	checkContent("src/app.js", `View source`)
	checkContent("src/app.js", `href="docs.html"`)
	checkContent("src/app.js", `https://github.com/xyzzyapps/goweb`)
	checkContent("src/app.js", `handleAdd`)
	checkContent("src/app.js", `handleToggle`)
	checkContent("src/app.js", `handleDelete`)
	checkContent("src/style.css", `--color-primary: #ffcd42`)
	checkContent("src/components/add-todo.js", `export function AddTodo`)
	checkContent("src/components/todo-list.js", `export function TodoList`)
	checkContent("src/components/todo-item.js", `export function TodoItem`)
	checkContent("LICENSE", `Creative Commons Attribution-ShareAlike`)
	checkContent("LICENSE", `Tester`)
	checkContent("LICENSE", `tester@example.com`)

	// Verify debug chunks are included when debug=true.
	checkContent("src/app.js", `console.log("[debug] added todo:"`)
	checkContent("src/app.js", `console.log("[debug] toggled todo:"`)
	checkContent("src/app.js", `console.log("[debug] deleted todo:"`)

	// Verify the override-demo tag is present in todo-list.
	// (todo-list.tsx has tags: component, override-demo)
	// Since we can't easily check tags on output files, just verify the file exists.
}
