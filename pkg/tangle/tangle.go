// Package tangle implements code extraction from goweb literate programming
// documents. It resolves <<ref>> references, applies pipe commands, exec
// commands, session-based execution, and writes the resulting source files.
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
	"sync"

	"github.com/manic/goweb/pkg/graph"
	"github.com/manic/goweb/pkg/parser"
)

// SessionManager manages persistent processes for session-based execution.
type SessionManager struct {
	mu    sync.Mutex
	procs map[string]*exec.Cmd
	stdin map[string]io.WriteCloser
}

// NewSessionManager creates a new session manager.
func NewSessionManager() *SessionManager {
	return &SessionManager{
		procs: make(map[string]*exec.Cmd),
		stdin: make(map[string]io.WriteCloser),
	}
}

// Exec runs body through the command identified by exeCmd.
// If session is non-empty, a persistent process is used and state carries over.
func (sm *SessionManager) Exec(exeCmd, session, body string, pipeDir string) (string, error) {
	if session == "" {
		// One-shot execution.
		return runCommand(exeCmd, body, pipeDir)
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	cmd, exists := sm.procs[session]
	if !exists {
		// Start a new persistent process.
		parts := splitCommand(exeCmd)
		if len(parts) == 0 {
			return body, nil
		}
		c := exec.Command(parts[0], parts[1:]...)
		if pipeDir != "" {
			c.Dir = pipeDir
		}

		stdin, err := c.StdinPipe()
		if err != nil {
			return "", fmt.Errorf("session %q: stdin pipe: %w", session, err)
		}
		var stdout bytes.Buffer
		c.Stdout = &stdout
		c.Stderr = &stdout

		if err := c.Start(); err != nil {
			return "", fmt.Errorf("session %q: start: %w", session, err)
		}

		sm.procs[session] = c
		sm.stdin[session] = stdin

		// Write body to stdin and close it.
		_, _ = io.WriteString(stdin, body)
		stdin.Close()

		// Wait for the process to finish.
		if err := c.Wait(); err != nil {
			return "", fmt.Errorf("session %q: wait: %w\noutput: %s", session, err, stdout.String())
		}

		return stdout.String(), nil
	}

	// Reuse existing session process.
	stdin := sm.stdin[session]
	_, _ = io.WriteString(stdin, body)
	stdin.Close()

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("session %q: wait: %w\noutput: %s", session, err, stdout.String())
	}

	delete(sm.procs, session)
	delete(sm.stdin, session)
	return stdout.String(), nil
}

// Close terminates all active sessions.
func (sm *SessionManager) Close() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	for name, cmd := range sm.procs {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		delete(sm.procs, name)
	}
	sm.stdin = make(map[string]io.WriteCloser)
}

// runCommand executes a command with body as stdin and returns stdout.
func runCommand(exeCmd, body, pipeDir string) (string, error) {
	parts := splitCommand(exeCmd)
	if len(parts) == 0 {
		return body, nil
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdin = strings.NewReader(body)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if pipeDir != "" {
		cmd.Dir = pipeDir
	}

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("command %q failed: %w\nstderr: %s", exeCmd, err, stderr.String())
	}

	return stdout.String(), nil
}

// Tangle holds configuration for the tangling process.
type Tangle struct {
	// PipeDir is the working directory for pipe/exec commands.
	PipeDir string

	// DryRun, if true, prints what would be written without actually writing.
	DryRun bool

	// Stdout is where chunk output goes when no file: attribute is set
	// or when a specific chunk is requested.
	Stdout io.Writer

	// sessions manages persistent processes for session-based execution.
	sessions *SessionManager
}

// New creates a new Tangle with default settings.
func New() *Tangle {
	return &Tangle{
		Stdout:   os.Stdout,
		sessions: NewSessionManager(),
	}
}

// execChunk runs the chunk body through the exec command, optionally using a session.
func (t *Tangle) execChunk(c *parser.Chunk, body string) (string, error) {
	if c.ExecCmd == "" {
		return body, nil
	}
	return t.sessions.Exec(c.ExecCmd, c.SessionName, body, t.PipeDir)
}

// Close terminates all active sessions.
func (t *Tangle) Close() {
	t.sessions.Close()
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

			// Apply exec command if specified (runs body through interpreter).
			executed, err := t.execChunk(c, resolved)
			if err != nil {
				return fmt.Errorf("executing chunk %q: %w", c.Name, err)
			}

			// Apply pipe command if specified (transforms output).
			output, err := t.applyPipe(c.Name, executed, c.PipeCmd)
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
