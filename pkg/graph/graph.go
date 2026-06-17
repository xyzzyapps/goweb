// Package graph builds a dependency graph from named chunks and provides
// topological ordering for tangle resolution.
package graph

import (
	"fmt"
	"strings"

	"github.com/manic/goweb/pkg/parser"
)

// Graph represents a chunk dependency graph.
type Graph struct {
	// nodes maps chunk name to its dependencies (references).
	nodes map[string]map[string]bool

	// chunks maps chunk name to the Chunk object.
	chunks map[string]*parser.Chunk

	// vars holds user-provided variables for {{var}} substitution.
	vars map[string]string
}

// New builds a dependency graph from a Document.
// It reads all chunk bodies and extracts <<ref>> references.
func New(doc *parser.Document) (*Graph, error) {
	g := &Graph{
		nodes:  make(map[string]map[string]bool),
		chunks: make(map[string]*parser.Chunk),
		vars:   make(map[string]string),
	}

	// Copy vars for local use.
	if doc.Vars != nil {
		for k, v := range doc.Vars {
			g.vars[k] = v
		}
	}

	for _, c := range doc.Chunks {
		if _, exists := g.chunks[c.Name]; exists {
			return nil, fmt.Errorf("duplicate chunk %q", c.Name)
		}
		g.chunks[c.Name] = c

		deps := findReferences(c.Body)
		g.nodes[c.Name] = deps
	}

	return g, nil
}

// findReferences extracts <<name>> references from a body string.
// It skips <<name>>= definitions.
func findReferences(body string) map[string]bool {
	refs := make(map[string]bool)
	i := 0
	for i < len(body) {
		start := strings.Index(body[i:], "<<")
		if start == -1 {
			break
		}
		start += i

		end := strings.Index(body[start+2:], ">>")
		if end == -1 {
			break
		}
		end += start + 2

		// Check if followed by '=' (definition, not reference).
		// end points to the first '>' of ">>", so we need end+2 for '='.
		if end+2 < len(body) && body[end+2] == '=' {
			i = end + 3
			continue
		}

		name := strings.TrimSpace(body[start+2 : end])
		if name != "" {
			refs[name] = true
		}
		i = end + 2
	}
	return refs
}

// TopologicalSort returns chunks in dependency-first order using Kahn's
// algorithm. Nodes with no dependencies come first; a node is emitted once
// all its dependencies have been emitted. Returns an error if a cycle is
// detected.
func (g *Graph) TopologicalSort() ([]*parser.Chunk, error) {
	// Build reverse map: dependency → list of chunks that depend on it.
	revDeps := make(map[string][]string)
	for name, deps := range g.nodes {
		if len(deps) == 0 {
			continue
		}
		for dep := range deps {
			revDeps[dep] = append(revDeps[dep], name)
		}
	}

	// out-degree = number of dependencies still to be resolved.
	outDegree := make(map[string]int)
	for name := range g.nodes {
		outDegree[name] = len(g.nodes[name])
	}

	// Seed queue with nodes that have zero dependencies.
	var queue []string
	for name := range g.nodes {
		if outDegree[name] == 0 {
			queue = append(queue, name)
		}
	}

	var sorted []*parser.Chunk
	processed := make(map[string]bool)

	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]

		if processed[name] {
			continue
		}
		processed[name] = true
		sorted = append(sorted, g.chunks[name])

		for _, dependent := range revDeps[name] {
			if !processed[dependent] {
				outDegree[dependent]--
				if outDegree[dependent] == 0 {
					queue = append(queue, dependent)
				}
			}
		}
	}

	if len(processed) != len(g.nodes) {
		var unprocessed []string
		for name := range g.nodes {
			if !processed[name] {
				unprocessed = append(unprocessed, name)
			}
		}
		return sorted, fmt.Errorf("circular dependency detected involving: %s",
			strings.Join(unprocessed, ", "))
	}

	return sorted, nil
}

// Resolve expands all references in a chunk body recursively.
// It returns the fully resolved body with all <<ref>> replaced by their content
// and {{var}} variables substituted.
// refs tracks the current expansion chain to detect circular references.
func (g *Graph) Resolve(name string, refs map[string]bool) (string, error) {
	if refs == nil {
		refs = make(map[string]bool)
	}

	chunk, ok := g.chunks[name]
	if !ok {
		return "", fmt.Errorf("undefined chunk %q", name)
	}

	if refs[name] {
		return "", fmt.Errorf("circular reference involving chunk %q", name)
	}
	refs[name] = true
	defer delete(refs, name)

	body := chunk.Body
	var result strings.Builder
	i := 0
	for i < len(body) {
		start := strings.Index(body[i:], "<<")
		if start == -1 {
			result.WriteString(body[i:])
			break
		}
		start += i
		result.WriteString(body[i:start])

		end := strings.Index(body[start+2:], ">>")
		if end == -1 {
			result.WriteString(body[start:])
			break
		}
		end += start + 2

		// Check if followed by '=' (definition, not reference).
		// end points to the first '>' of ">>", so check end+2 for '='.
		if end+2 < len(body) && body[end+2] == '=' {
			result.WriteString(body[start : end+3])
			i = end + 3
			continue
		}

		refName := strings.TrimSpace(body[start+2 : end])
		if refName == "" {
			result.WriteString("<<>>")
			i = end + 2
			continue
		}

		resolved, err := g.Resolve(refName, refs)
		if err != nil {
			return "", fmt.Errorf("resolving <%s> in chunk %q: %w", refName, name, err)
		}
		result.WriteString(resolved)
		i = end + 2
	}

	// Substitute {{var}} variables in the resolved result.
	return g.expandVars(result.String()), nil
}

// expandVars replaces {{name}} with the corresponding variable value.
// Unknown variables are left as-is.
func (g *Graph) expandVars(s string) string {
	var result strings.Builder
	i := 0
	for i < len(s) {
		start := strings.Index(s[i:], "{{")
		if start == -1 {
			result.WriteString(s[i:])
			break
		}
		start += i
		result.WriteString(s[i:start])

		end := strings.Index(s[start+2:], "}}")
		if end == -1 {
			result.WriteString(s[start:])
			break
		}
		end += start + 2

		varName := strings.TrimSpace(s[start+2 : end])
		if varName == "" {
			result.WriteString("{{}}")
			i = end + 2
			continue
		}

		if val, ok := g.vars[varName]; ok {
			result.WriteString(val)
		} else {
			// Variable not set — leave the placeholder as-is.
			result.WriteString(s[start : end+2])
		}
		i = end + 2
	}
	return result.String()
}
