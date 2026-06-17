// Package weave generates clean markdown documentation from goweb literate
// programming sources. It strips goweb control syntax (<<>> directives,
// chunk definitions, terminators) and produces readable markdown.
package weave

import (
	"fmt"
	"io"
	"strings"

	"github.com/manic/goweb/pkg/preproc"
)

// Weave generates clean markdown from a goweb source file.
// It first runs the preprocessor to resolve imports and conditionals,
// then strips remaining goweb control syntax.
func Weave(sourcePath string, vars map[string]string, w io.Writer) error {
	// Preprocess: resolve imports and conditionals.
	result, err := preproc.Process(sourcePath, vars)
	if err != nil {
		return fmt.Errorf("preprocessing %s: %w", sourcePath, err)
	}

	// Strip goweb control syntax from the preprocessed lines.
	cleaned := stripControlSyntax(result.Lines)

	_, err = io.WriteString(w, strings.Join(cleaned, "\n"))
	return err
}

// stripControlSyntax removes goweb control lines and markers from a list of
// source lines, leaving clean markdown.
func stripControlSyntax(lines []string) []string {
	var out []string
	inFence := false
	fenceChar := ""

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Track fenced code blocks so we know when we're inside code.
		if isFenceLine(trimmed) {
			fc := getFenceChars(trimmed)
			if !inFence {
				inFence = true
				fenceChar = fc
				// Keep fence opening line but strip any goweb attributes.
				out = append(out, cleanFenceOpen(line))
				continue
			} else if len(fc) >= 3 && fc[0] == fenceChar[0] && len(fc) >= len(fenceChar) {
				inFence = false
				fenceChar = ""
				out = append(out, line)
				continue
			}
		}

		if inFence {
			// Inside a code block.
			// Strip chunk definition headers.
			if isChunkDefStart(trimmed) {
				continue // Skip the <<name>>= line.
			}
			// Strip chunk definition terminators (>> on its own line).
			if trimmed == ">>" {
				continue
			}
			// Remove any remaining <<ref>> references or directives.
			cleaned := removeAngleBrackets(line)
			out = append(out, cleaned)
		} else {
			// Outside code blocks.
			// Skip any stray goweb directives that weren't caught by preprocessor.
			if isGowebDirective(trimmed) {
				continue
			}
			// Strip any stray >> on its own line.
			if trimmed == ">>" {
				continue
			}
			// Skip chunk definition starts.
			if isChunkDefStart(trimmed) {
				continue
			}
			// Also clean any inline <<...>> that might remain.
			cleaned := removeAngleBrackets(line)
			out = append(out, cleaned)
		}
	}

	return out
}

// isFenceLine checks if a line looks like a fenced code block marker.
func isFenceLine(trimmed string) bool {
	if len(trimmed) < 3 {
		return false
	}
	first := trimmed[0]
	if first != '`' && first != '~' {
		return false
	}
	count := 0
	for _, ch := range trimmed {
		if byte(ch) != first {
			break
		}
		count++
	}
	return count >= 3
}

// getFenceChars returns the fence character sequence from a code fence line.
func getFenceChars(trimmed string) string {
	if len(trimmed) < 3 {
		return ""
	}
	first := trimmed[0]
	count := 0
	for _, ch := range trimmed {
		if byte(ch) != first {
			break
		}
		count++
	}
	if count >= 3 {
		return trimmed[:count]
	}
	return ""
}

// cleanFenceOpen removes goweb attributes from a fenced code block opening line.
func cleanFenceOpen(line string) string {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) < 3 {
		return line
	}
	first := trimmed[0]
	count := 0
	for _, ch := range trimmed {
		if byte(ch) != first {
			break
		}
		count++
	}
	if count < 3 {
		return line
	}
	// Keep only the fence chars and the language tag (before any {).
	rest := strings.TrimSpace(trimmed[count:])
	// Find first { or space
	end := len(rest)
	for i := 0; i < len(rest); i++ {
		if rest[i] == '{' || rest[i] == ' ' || rest[i] == '\t' {
			end = i
			break
		}
	}
	lang := rest[:end]
	if lang != "" {
		return trimmed[:count] + lang
	}
	return trimmed[:count]
}

// isChunkDefStart checks if a line starts a chunk definition (<<name>>=).
func isChunkDefStart(trimmed string) bool {
	return strings.HasPrefix(trimmed, "<<") && strings.Contains(trimmed, ">>=")
}

// isGowebDirective checks if a trimmed line is a goweb control directive.
func isGowebDirective(trimmed string) bool {
	return strings.HasPrefix(trimmed, "<<if ") ||
		strings.HasPrefix(trimmed, "<<elif ") ||
		trimmed == "<<else>>" ||
		trimmed == "<<end>>" ||
		strings.HasPrefix(trimmed, "<<import")
}

// removeAngleBrackets removes or neutralizes <<...>> sequences from a line.
// It replaces them with the inner content (for references) or removes them.
func removeAngleBrackets(line string) string {
	var result strings.Builder
	i := 0
	for i < len(line) {
		start := strings.Index(line[i:], "<<")
		if start == -1 {
			result.WriteString(line[i:])
			break
		}
		start += i
		result.WriteString(line[i:start])

		end := strings.Index(line[start+2:], ">>")
		if end == -1 {
			result.WriteString(line[start:])
			break
		}
		end += start + 2

		inner := strings.TrimSpace(line[start+2 : end])

		// Check if it's a definition (<<name>>=)
		// end points to the first '>' of ">>", so check end+2 for '='.
		if end+2 < len(line) && line[end+2] == '=' {
			// Skip the entire definition marker.
			i = end + 3
			continue
		}

		// If inner is non-empty, it's a reference — keep the inner text.
		if inner != "" {
			result.WriteString(inner)
		}
		i = end + 2
	}
	return result.String()
}
