// Package tangle implements code extraction from goweb literate programming
// documents. It resolves <<ref>> references, applies pipe commands, and
// writes the resulting source files.
package tangle

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/manic/goweb/pkg/graph"
	"github.com/manic/goweb/pkg/parser"
)

// Tangle holds configuration for the tangling process.
type Tangle struct {
	// PipeDir is the working directory for pipe commands.
	PipeDir string

	// DryRun, if true, prints what would be written without actually writing.
	DryRun bool

	// Stdout is where chunk output goes when no file: attribute is set
	// or when a specific chunk is requested.
	Stdout io.Writer
}

// New creates a new Tangle with default settings.
func New() *Tangle {
	return &Tangle{
		Stdout: os.Stdout,
	}
}

// Tangle processes a Document, resolving all references, applying pipes,
// and writing output files.
func (t *Tangle) Tangle(doc *parser.Document) error {
	g, err := graph.New(doc)
	if err != nil {
		return fmt.Errorf("building dependency graph: %w", err)
	}

	// Validate all references resolve.
	if err := parser.ValidateDocument(doc); err != nil {
		return err
	}

	// Group chunks by output file.
	byFile := doc.ChunksByFile()

	// Process files in sorted order for deterministic output.
	var files []string
	for f := range byFile {
		if f != "" {
			files = append(files, f)
		}
	}
	sort.Strings(files)

	for _, filename := range files {
		chunks := byFile[filename]
		var content strings.Builder

		for _, c := range chunks {
			resolved, err := g.Resolve(c.Name, nil)
			if err != nil {
				return fmt.Errorf("resolving chunk %q: %w", c.Name, err)
			}

			// Apply pipe command if specified.
			output, err := t.applyPipe(c.Name, resolved, c.PipeCmd)
			if err != nil {
				return fmt.Errorf("piping chunk %q: %w", c.Name, err)
			}

			content.WriteString(output)
			// Ensure trailing newline.
			if !strings.HasSuffix(output, "\n") {
				content.WriteString("\n")
			}
		}

		if t.DryRun {
			fmt.Fprintf(t.Stdout, "=== would write %s (%d bytes) ===\n%s\n",
				filename, content.Len(), content.String())
		} else {
			if err := writeFile(filename, content.String()); err != nil {
				return fmt.Errorf("writing %s: %w", filename, err)
			}
			fmt.Fprintf(t.Stdout, "wrote %s\n", filename)
		}
	}

	// Also process chunks without a file: attribute.
	if orphans, ok := byFile[""]; ok && len(orphans) > 0 {
		if t.DryRun {
			fmt.Fprintln(t.Stdout, "=== chunks without file: attribute ===")
		}
		for _, c := range orphans {
			resolved, err := g.Resolve(c.Name, nil)
			if err != nil {
				return fmt.Errorf("resolving chunk %q: %w", c.Name, err)
			}
			output, err := t.applyPipe(c.Name, resolved, c.PipeCmd)
			if err != nil {
				return fmt.Errorf("piping chunk %q: %w", c.Name, err)
			}
			if t.DryRun {
				fmt.Fprintf(t.Stdout, "--- %s ---\n%s\n", c.Name, output)
			}
		}
	}

	return nil
}

// TangleChunk resolves a single named chunk and writes it to the writer.
func (t *Tangle) TangleChunk(doc *parser.Document, name string, w io.Writer) error {
	g, err := graph.New(doc)
	if err != nil {
		return fmt.Errorf("building dependency graph: %w", err)
	}

	// Check if the chunk exists.
	if _, err := g.Resolve(name, nil); err != nil {
		return fmt.Errorf("resolving chunk %q: %w", name, err)
	}

	resolved, err := g.Resolve(name, nil)
	if err != nil {
		return err
	}

	// Apply pipe if specified.
	chunk := findChunk(doc, name)
	var pipeCmd string
	if chunk != nil {
		pipeCmd = chunk.PipeCmd
	}

	output, err := t.applyPipe(name, resolved, pipeCmd)
	if err != nil {
		return fmt.Errorf("piping chunk %q: %w", name, err)
	}

	_, err = io.WriteString(w, output)
	return err
}

// findChunk finds a chunk by name in a document.
func findChunk(doc *parser.Document, name string) *parser.Chunk {
	for _, c := range doc.Chunks {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// applyPipe runs the chunk body through a pipe command if one is specified.
func (t *Tangle) applyPipe(name, body, pipeCmd string) (string, error) {
	if pipeCmd == "" {
		return body, nil
	}

	// Parse the pipe command.
	parts := splitCommand(pipeCmd)
	if len(parts) == 0 {
		return body, nil
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdin = strings.NewReader(body)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if t.PipeDir != "" {
		cmd.Dir = t.PipeDir
	}

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pipe command %q failed for chunk %q: %w\nstderr: %s",
			pipeCmd, name, err, stderr.String())
	}

	return stdout.String(), nil
}

// splitCommand splits a command string into parts, handling quoted strings.
func splitCommand(cmd string) []string {
	var parts []string
	var current strings.Builder
	inQuote := false

	for i := 0; i < len(cmd); i++ {
		c := cmd[i]
		switch {
		case c == '"':
			inQuote = !inQuote
		case c == ' ' && !inQuote:
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(c)
		}
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// writeFile writes content to a file, creating parent directories as needed.
func writeFile(path, content string) error {
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}
	return os.WriteFile(path, []byte(content), 0644)
}
