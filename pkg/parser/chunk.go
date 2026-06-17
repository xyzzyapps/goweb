// Package parser provides data structures and parsing for goweb literate
// programming markdown files.
package parser

import (
	"fmt"
	"sort"
)

// Chunk represents a named code chunk in a literate program.
// Chunks are defined via <<name>>= ... >> syntax inside fenced code blocks.
type Chunk struct {
	// Name is the chunk identifier (e.g., "package", "main").
	Name string

	// Language is the programming language from the fenced code block info string.
	Language string

	// Body is the raw chunk body, which may contain <<ref>> placeholders.
	Body string

	// File is the output file path from the file: attribute.
	// Empty means this chunk is not directly written to a file.
	File string

	// PipeCmd is the pipe command from the pipe: attribute.
	// Empty means no pipe transformation.
	PipeCmd string

	// Override, if true, means this chunk replaces (rather than appends to)
	// any previous definition with the same name.
	Override bool

	// Line is the 1-based source line where this chunk was defined.
	Line int

	// Source is the path of the source file this chunk came from.
	Source string
}

// Merge combines two chunks with the same name.
// If the new chunk has Override=true, it replaces the existing one entirely.
// Otherwise, the bodies are concatenated with a newline separator.
// Attributes from the first definition are kept.
func (c *Chunk) Merge(other *Chunk) {
	if other.Override {
		c.Body = other.Body
		c.Line = other.Line
		c.Source = other.Source
		c.Language = other.Language
		return
	}
	if c.Body != "" && other.Body != "" {
		c.Body += "\n" + other.Body
	} else if other.Body != "" {
		c.Body = other.Body
	}
	// Keep language from the first definition unless it was empty
	if c.Language == "" && other.Language != "" {
		c.Language = other.Language
	}
}

// Document represents the full result of parsing one or more literate
// programming source files.
type Document struct {
	// Chunks is the flat list of all named chunks across all source files.
	Chunks []*Chunk

	// Vars is the set of user-provided variables from --var flags.
	Vars map[string]string

	// Source is the primary source file path.
	Source string
}

// ChunkTable builds a map from chunk name to Chunk for fast lookup.
// If there are duplicate names across files, the first one wins and
// duplicates are reported.
func (doc *Document) ChunkTable() (map[string]*Chunk, error) {
	table := make(map[string]*Chunk, len(doc.Chunks))
	for _, c := range doc.Chunks {
		if existing, ok := table[c.Name]; ok {
			return nil, fmt.Errorf(
				"duplicate chunk %q: first defined in %s:%d, then in %s:%d",
				c.Name, existing.Source, existing.Line, c.Source, c.Line,
			)
		}
		table[c.Name] = c
	}
	return table, nil
}

// ChunksByFile groups chunks by their output file.
// Chunks without a file: attribute are grouped under the key "".
func (doc *Document) ChunksByFile() map[string][]*Chunk {
	grouped := make(map[string][]*Chunk)
	for _, c := range doc.Chunks {
		key := c.File
		if key == "" {
			key = ""
		}
		grouped[key] = append(grouped[key], c)
	}
	// Sort within each group by source position for deterministic output.
	for _, group := range grouped {
		sort.Slice(group, func(i, j int) bool {
			if group[i].Source != group[j].Source {
				return group[i].Source < group[j].Source
			}
			return group[i].Line < group[j].Line
		})
	}
	return grouped
}

// AllReferences returns the set of unique reference names found in all chunks.
func (doc *Document) AllReferences() map[string]bool {
	refs := make(map[string]bool)
	for _, c := range doc.Chunks {
		for _, ref := range extractReferences(c.Body) {
			refs[ref] = true
		}
	}
	return refs
}

// extractReferences finds all <<name>> references in a body string.
// It does not match <<name>>= (definitions) or other directives.
func extractReferences(body string) []string {
	var refs []string
	i := 0
	for i < len(body) {
		// Look for <<
		start := findByte(body, '<', i)
		if start == -1 || start+1 >= len(body) || body[start+1] != '<' {
			break
		}
		// Check if this is <<name>>= (definition) — skip those.
		end := findByte(body, '>', start+2)
		if end == -1 || end+1 >= len(body) || body[end+1] != '>' {
			i = start + 2
			continue
		}
		// end points to first '>' of ">>", check end+2 for '='.
		if end+2 < len(body) && body[end+2] == '=' {
			i = end + 3
			continue
		}
		name := body[start+2 : end]
		if name != "" {
			refs = append(refs, name)
		}
		i = end + 2
	}
	return refs
}

// findByte finds the next occurrence of b in s starting from start.
func findByte(s string, b byte, start int) int {
	for i := start; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}
