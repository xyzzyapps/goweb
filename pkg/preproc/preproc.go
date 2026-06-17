// Package preproc handles goweb preprocessor directives:
//   - <<import "path">>  — load chunks from another file
//   - <<if expr>> / <<elif expr>> / <<else>> / <<end>> — conditional blocks
//
// It reads source files line by line, follows imports recursively, evaluates
// conditionals using user-provided variables, and returns filtered lines.
package preproc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// State tracks the preprocessor state machine for conditionals.
type State int

const (
	StateNormal   State = iota // Outside any conditional
	StateActive                // Inside an active (taken) branch
	StateSkipping              // Inside a skipped branch, looking for else/end
	StateDone                  // Inside a taken branch after else — skip remaining elif/else
)

// Result holds the output of preprocessing.
type Result struct {
	// Lines is the filtered, flattened list of source lines.
	Lines []string

	// Sources is the set of all source files that were loaded.
	Sources map[string]bool
}

// Process reads a root source file, follows <<import>> directives recursively,
// evaluates <<if>>/<<elif>>/<<else>>/<<end>> conditionals using vars,
// and returns the filtered lines.
func Process(rootPath string, vars map[string]string) (*Result, error) {
	result := &Result{
		Sources: make(map[string]bool),
	}
	absRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, fmt.Errorf("resolving path %s: %w", rootPath, err)
	}
	if err := processFileWithImports(absRoot, vars, result); err != nil {
		return nil, err
	}
	return result, nil
}

// processFileWithImports reads a file and processes imports inline.
func processFileWithImports(path string, vars map[string]string, result *Result) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolving path %s: %w", path, err)
	}
	if result.Sources[absPath] {
		return nil
	}
	result.Sources[absPath] = true

	data, err := os.ReadFile(absPath)
	if err != nil {
		return fmt.Errorf("reading %s: %w", absPath, err)
	}

	// Strip YAML frontmatter if present.
	content := stripFrontmatter(string(data), vars)

	rawLines := splitLines(content)
	dir := filepath.Dir(absPath)

	// State machine for conditionals.
	type branch struct {
		state   State
		hasElse bool
	}
	stack := []branch{{state: StateNormal}}

	var outLines []string

	// flushOutLines writes pending outLines to result.Lines and resets.
	flushOutLines := func() {
		if len(outLines) > 0 {
			result.Lines = append(result.Lines, outLines...)
			outLines = nil
		}
	}

	for i, line := range rawLines {
		trimmed := strings.TrimSpace(line)
		currentState := stack[len(stack)-1].state

		// Check for import or override directive.
		if isImportDirective(trimmed) || isOverrideDirective(trimmed) {
			isOverride := isOverrideDirective(trimmed)
			importPath := parseImportPath(trimmed)
			resolved := importPath
			if !filepath.IsAbs(resolved) {
				resolved = filepath.Join(dir, resolved)
			}
			// Flush lines accumulated so far before the imported content.
			flushOutLines()
			if isOverride {
				// Inject marker so the parser knows subsequent chunks override.
				result.Lines = append(result.Lines, "<<__goweb_override__>>")
			}
			if err := processFileWithImports(resolved, vars, result); err != nil {
				return fmt.Errorf("importing %s from %s line %d: %w", importPath, absPath, i+1, err)
			}
			continue
		}

		// Check for conditional directives.
		if isConditionalDirective(trimmed) {
			dirType, expr := parseConditionalDirective(trimmed)
			switch dirType {
			case "if":
				if currentState == StateNormal {
					if evaluateCondition(expr, vars) {
						stack = append(stack, branch{state: StateActive})
					} else {
						stack = append(stack, branch{state: StateSkipping})
					}
				} else {
					// Nested inside skipped or done — skip the whole nested block.
					stack = append(stack, branch{state: StateSkipping})
				}
			case "elif":
				if len(stack) < 2 {
					return fmt.Errorf("%s line %d: <<elif>> without matching <<if>>", absPath, i+1)
				}
				top := &stack[len(stack)-1]
				if top.hasElse {
					return fmt.Errorf("%s line %d: <<elif>> after <<else>>", absPath, i+1)
				}
				switch top.state {
				case StateActive, StateDone:
					top.state = StateSkipping
				case StateSkipping:
					if evaluateCondition(expr, vars) {
						top.state = StateActive
					}
				}
			case "else":
				if len(stack) < 2 {
					return fmt.Errorf("%s line %d: <<else>> without matching <<if>>", absPath, i+1)
				}
				top := &stack[len(stack)-1]
				if top.hasElse {
					return fmt.Errorf("%s line %d: duplicate <<else>>", absPath, i+1)
				}
				top.hasElse = true
				switch top.state {
				case StateActive:
					top.state = StateDone
				case StateSkipping:
					top.state = StateActive
				default:
					top.state = StateSkipping
				}
			case "end":
				if len(stack) < 2 {
					return fmt.Errorf("%s line %d: <<end>> without matching <<if>>", absPath, i+1)
				}
				stack = stack[:len(stack)-1]
			}
			continue
		}

		// Regular line — output if we're in an active state.
		if currentState == StateNormal || currentState == StateActive {
			outLines = append(outLines, line)
		}
	}

	if len(stack) != 1 {
		return fmt.Errorf("%s: unclosed <<if>> block (stack depth %d)", absPath, len(stack))
	}

	result.Lines = append(result.Lines, outLines...)
	return nil
}

// isImportDirective checks if a trimmed line is an <<import "path">> directive.
func isImportDirective(line string) bool {
	if !strings.HasPrefix(line, "<<import") {
		return false
	}
	if len(line) < 10 {
		return false
	}
	ch := line[8]
	return ch == ' ' || ch == '"'
}

// isOverrideDirective checks if a trimmed line is an <<override "path">> directive.
func isOverrideDirective(line string) bool {
	if !strings.HasPrefix(line, "<<override") {
		return false
	}
	if len(line) < 11 {
		return false
	}
	ch := line[10]
	return ch == ' ' || ch == '"'
}

// parseImportPath extracts the path from <<import "path">> or <<override "path">>.
func parseImportPath(line string) string {
	// Line is like: <<import "path/to/file.md">>
	start := strings.Index(line, "\"")
	if start == -1 {
		return ""
	}
	end := strings.LastIndex(line, "\"")
	if end <= start {
		return ""
	}
	return line[start+1 : end]
}

// isConditionalDirective checks if a trimmed line is a conditional directive.
func isConditionalDirective(line string) bool {
	return strings.HasPrefix(line, "<<if ") ||
		strings.HasPrefix(line, "<<elif ") ||
		line == "<<else>>" ||
		line == "<<end>>"
}

// parseConditionalDirective returns the directive type and expression.
func parseConditionalDirective(line string) (dirType string, expr string) {
	if strings.HasPrefix(line, "<<if ") {
		inner := strings.TrimPrefix(line, "<<if ")
		expr = strings.TrimSuffix(inner, ">>")
		return "if", strings.TrimSpace(expr)
	}
	if strings.HasPrefix(line, "<<elif ") {
		inner := strings.TrimPrefix(line, "<<elif ")
		expr = strings.TrimSuffix(inner, ">>")
		return "elif", strings.TrimSpace(expr)
	}
	if line == "<<else>>" {
		return "else", ""
	}
	if line == "<<end>>" {
		return "end", ""
	}
	return "", ""
}

// splitLines splits content into lines, handling both \r\n and \n,
// and discards a trailing empty line from a final newline.
func splitLines(content string) []string {
	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	// Discard trailing empty line from final newline.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// stripFrontmatter detects and removes YAML frontmatter (content between --- markers)
// from the beginning of a file. It also extracts any "vars:" values into the vars map.
func stripFrontmatter(content string, vars map[string]string) string {
	lines := strings.Split(content, "\n")
	if len(lines) < 2 || strings.TrimSpace(lines[0]) != "---" {
		return content
	}

	// Find closing ---
	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			endIdx = i
			break
		}
	}
	if endIdx == -1 {
		return content
	}

	// Parse simple key: value pairs from frontmatter.
	for i := 1; i < endIdx; i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		colonIdx := strings.Index(line, ":")
		if colonIdx == -1 {
			continue
		}
		key := strings.TrimSpace(line[:colonIdx])
		val := strings.TrimSpace(line[colonIdx+1:])
		// Remove quotes.
		val = strings.Trim(val, "\"'")
		if key != "" && val != "" {
			// If key is "vars", parse sub-keys.
			if key == "vars" {
				// Value is the rest of the line, could be inline vars.
				continue
			}
			// Only set if not already set (CLI flags take precedence).
			if _, exists := vars[key]; !exists {
				vars[key] = val
			}
		}
	}

	// Return content after frontmatter.
	return strings.Join(lines[endIdx+1:], "\n")
}

// evaluateCondition evaluates a simple conditional expression.
func evaluateCondition(expr string, vars map[string]string) bool {
	expr = strings.TrimSpace(expr)

	// Check for != first (before == to avoid partial match).
	if strings.Contains(expr, "!=") {
		parts := strings.SplitN(expr, "!=", 2)
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		// Remove surrounding quotes if present.
		val = strings.Trim(val, "\"")
		actual, ok := vars[key]
		if !ok {
			return val != "" // If key doesn't exist, true only if comparing to empty
		}
		return actual != val
	}

	// Check for ==
	if strings.Contains(expr, "==") {
		parts := strings.SplitN(expr, "==", 2)
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, "\"")
		actual, ok := vars[key]
		if !ok {
			return val == "" // If key doesn't exist, true only if comparing to empty
		}
		return actual == val
	}

	// Simple variable name — true if the var is set and not empty/"false"/"0".
	val, ok := vars[expr]
	if !ok {
		return false
	}
	val = strings.ToLower(strings.TrimSpace(val))
	return val != "" && val != "false" && val != "0" && val != "no"
}
