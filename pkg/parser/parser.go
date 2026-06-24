package parser

import (
	"fmt"
	"strings"
)

// fencedBlock tracks the state of parsing a fenced code block.
type fencedBlock struct {
	language string
	lines    []string
	start    int // 1-based line number in source
}

// ParseLines extracts chunks from preprocessed markdown lines.
// It identifies fenced code blocks and extracts <<name>>= ... >> chunk
// definitions from within them. Chunk definitions can also appear outside
// fenced code blocks (true noweb-style).
func ParseLines(lines []string, sourcePath string) (*Document, error) {
	doc := &Document{
		Chunks: nil,
		Source: sourcePath,
		Vars:   make(map[string]string),
	}

	var chunks []*Chunk
	var currentFence *fencedBlock
	var fenceChar string          // ``` or ~~~
	var chunkDef *chunkDefinition // non-nil while inside a chunk definition
	overrideMode := false         // set by <<__goweb_override__>> marker

	for i, line := range lines {
		lineNum := i + 1 // 1-based

		// If we're inside a chunk definition (outside a fence), handle it.
		if chunkDef != nil && currentFence == nil {
			kind, literal := terminatorKind(line)
			if kind == 1 {
				// End of chunk definition.
			chunk := &Chunk{
				Name:        chunkDef.name,
				Language:    chunkDef.language,
				Body:        strings.Join(chunkDef.body, "\n"),
				File:        chunkDef.file,
				PipeCmd:     chunkDef.pipeCmd,
				ExecCmd:     chunkDef.execCmd,
				SessionName: chunkDef.sessionName,
				Line:        chunkDef.startLine,
				Source:      sourcePath,
				Override:    chunkDef.override,
			}
			if chunkDef.tags != "" {
				for _, tag := range strings.Split(chunkDef.tags, ",") {
					trimmed := strings.TrimSpace(tag)
					if trimmed != "" {
						chunk.Tags = append(chunk.Tags, trimmed)
					}
				}
			}
			chunks = addChunk(chunks, chunk)
				chunkDef = nil
			} else if kind >= 2 {
				chunkDef.body = append(chunkDef.body, literal)
			} else {
				chunkDef.body = append(chunkDef.body, line)
			}
			continue
		}

		// Check for fenced code block start/end.
		if isFenceOpen(line) {
			if currentFence != nil {
				// Close current fence (handles edge case of unclosed fence).
				for _, c := range extractChunksFromFence(currentFence, sourcePath) {
					if overrideMode {
						c.Override = true
					}
					chunks = addChunk(chunks, c)
				}
				currentFence = nil
			}
			fenceChar = getFenceString(line)
			lang := parseFenceLanguage(line)
			currentFence = &fencedBlock{
				language: lang,
				start:    lineNum,
			}
			continue
		}

		if currentFence != nil && isFenceClose(line, fenceChar) {
			for _, c := range extractChunksFromFence(currentFence, sourcePath) {
				if overrideMode {
					c.Override = true
				}
				chunks = addChunk(chunks, c)
			}
			currentFence = nil
			fenceChar = ""
			continue
		}

		if currentFence != nil {
			currentFence.lines = append(currentFence.lines, line)
			continue
		}

		// Handle override marker.
		if strings.TrimSpace(line) == "<<__goweb_override__>>" {
			overrideMode = true
			continue
		}

		// Outside a fence — check for chunk definition starts.
		trimmed := strings.TrimSpace(line)
		if isChunkDefinitionStart(trimmed) {
			name, attrs := parseChunkHeader(trimmed)
			chunkDef = &chunkDefinition{
				name:      name,
				startLine: lineNum,
			}
			for k, v := range attrs {
				switch k {
				case "file":
					chunkDef.file = v
				case "pipe":
					chunkDef.pipeCmd = v
				case "exec":
					chunkDef.execCmd = v
				case "session":
					chunkDef.sessionName = v
				case "tags":
					chunkDef.tags = v
				case "override":
					chunkDef.override = v == "true"
				}
			}
			// If overrideMode is active, this chunk overrides any existing.
			if overrideMode {
				chunkDef.override = true
			}
			continue
		}
	}

	// Handle unclosed fence at EOF.
	if currentFence != nil {
		for _, c := range extractChunksFromFence(currentFence, sourcePath) {
			if overrideMode {
				c.Override = true
			}
			chunks = addChunk(chunks, c)
		}
	}

	// Handle unclosed chunk definition at EOF.
	if chunkDef != nil {
		chunk := &Chunk{
			Name:     chunkDef.name,
			Body:     strings.Join(chunkDef.body, "\n"),
			Line:     chunkDef.startLine,
			Source:   sourcePath,
			Override: chunkDef.override,
		}
		chunk.File = chunkDef.file
		chunk.PipeCmd = chunkDef.pipeCmd
		chunks = addChunk(chunks, chunk)
	}

	// Auto-detect language from file: extension for chunks without a language.
	inferLanguages(chunks)

	doc.Chunks = chunks
	return doc, nil
}

// chunkDefinition tracks state while parsing a chunk definition outside a fence.
type chunkDefinition struct {
	name        string
	language    string
	body        []string
	file        string
	pipeCmd     string
	execCmd     string
	sessionName string
	tags        string // comma-separated
	override    bool
	startLine   int
}

// isFenceOpen checks if a line starts a fenced code block.
func isFenceOpen(line string) bool {
	trimmed := strings.TrimSpace(line)
	if len(trimmed) < 3 {
		return false
	}
	// Check for ``` or ~~~
	first := trimmed[0]
	if first != '`' && first != '~' {
		return false
	}
	for _, ch := range trimmed {
		if byte(ch) != first {
			// We allow other characters after the fence (like the language tag).
			// But all fence chars must be contiguous at the start.
			break
		}
		// Count consecutive fence chars.
	}
	// Check that the first 3 chars are fence chars.
	for i := 0; i < 3 && i < len(trimmed); i++ {
		if trimmed[i] != first {
			return false
		}
	}
	// Must be 3 or more fence chars.
	return true
}

// getFenceString returns the leading fence characters (``` or ~~~).
func getFenceString(line string) string {
	trimmed := strings.TrimSpace(line)
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

// isFenceClose checks if a line closes a fenced code block.
func isFenceClose(line string, fenceChars string) bool {
	if fenceChars == "" {
		return false
	}
	trimmed := strings.TrimSpace(line)
	if len(trimmed) < len(fenceChars) {
		return false
	}
	// The line must consist of the same fence character repeated,
	// optionally followed by whitespace.
	for i, ch := range trimmed {
		if i < len(fenceChars) {
			if byte(ch) != fenceChars[0] {
				return false
			}
		} else {
			if ch != ' ' && ch != '\t' {
				return false
			}
		}
	}
	return true
}

// parseFenceLanguage extracts the language tag from a fenced code block opening line.
// E.g., "```go {chunk=\"main\"}" -> "go"
func parseFenceLanguage(line string) string {
	trimmed := strings.TrimSpace(line)
	// Skip the fence characters.
	i := 0
	for i < len(trimmed) && (trimmed[i] == '`' || trimmed[i] == '~') {
		i++
	}
	rest := strings.TrimSpace(trimmed[i:])
	// Take everything before { or first space, whichever comes first.
	end := len(rest)
	for j := 0; j < len(rest); j++ {
		if rest[j] == ' ' || rest[j] == '\t' || rest[j] == '{' {
			end = j
			break
		}
	}
	return strings.TrimSpace(rest[:end])
}

// terminatorKind classifies a line in a chunk body.
//
//	0 = not a terminator (keep as body)
//	1 = bare ">>" (terminate chunk)
//	2 = escaped "\>>" (literal ">>", continue)
//	3 = double-escaped "\\>>" (literal "\>>", continue)
func terminatorKind(line string) (kind int, literal string) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "\\\\>>" {
		return 3, "\\\\>>"
	}
	if trimmed == "\\>>" {
		return 2, ">>"
	}
	if trimmed == ">>" {
		return 1, ""
	}
	return 0, line
}

// addChunk adds a chunk to the list, merging with any existing chunk
// that has the same name. If the new chunk has Override=true, it replaces
// the existing one. Otherwise bodies are concatenated.
func addChunk(chunks []*Chunk, new *Chunk) []*Chunk {
	for _, existing := range chunks {
		if existing.Name == new.Name {
			existing.Merge(new)
			return chunks
		}
	}
	return append(chunks, new)
}

// extractChunksFromFence extracts chunks from lines within a fenced code block.
// It looks for <<name>>= ... >> patterns.
func extractChunksFromFence(fence *fencedBlock, sourcePath string) []*Chunk {
	var chunks []*Chunk
	lines := fence.lines
	baseLine := fence.start // The fence opening line; first code line is baseLine+1

	i := 0
	for i < len(lines) {
		line := lines[i]
		// Look for <<name>>= (possibly with attributes like file:, pipe:)
		if trimmed := strings.TrimSpace(line); isChunkDefinitionStart(trimmed) {
			name, attrs := parseChunkHeader(trimmed)
			bodyStart := baseLine + i + 1

			// Collect body lines until >> on its own line.
			// (escaped \>> and \\>> produce >> and \>> literal)
			var bodyLines []string
			i++
			for i < len(lines) {
				kind, literal := terminatorKind(lines[i])
				if kind == 1 {
					i++
					break
				}
				if kind >= 2 {
					bodyLines = append(bodyLines, literal)
				} else {
					bodyLines = append(bodyLines, lines[i])
				}
				i++
			}

			chunk := &Chunk{
				Name:     name,
				Language: fence.language,
				Body:     strings.Join(bodyLines, "\n"),
				Line:     bodyStart,
				Source:   sourcePath,
			}

			// Apply attributes from the header.
			for k, v := range attrs {
				switch k {
				case "file":
					chunk.File = v
				case "pipe":
					chunk.PipeCmd = v
				case "exec":
					chunk.ExecCmd = v
				case "session":
					chunk.SessionName = v
				case "tags":
					for _, tag := range strings.Split(v, ",") {
						trimmed := strings.TrimSpace(tag)
						if trimmed != "" {
							chunk.Tags = append(chunk.Tags, trimmed)
						}
					}
				case "override":
					chunk.Override = v == "true"
				}
			}

			chunks = append(chunks, chunk)
		} else {
			i++
		}
	}

	return chunks
}

// isChunkDefinitionStart checks if a trimmed line starts a chunk definition:
// <<name>>= optionally followed by attributes like file:, pipe:
func isChunkDefinitionStart(line string) bool {
	return strings.HasPrefix(line, "<<") && strings.Contains(line, ">>=")
}

// parseChunkHeader parses a <<name>>= file: path pipe: cmd line.
// Also supports the bare keyword "override" (no colon).
// Returns the chunk name and a map of attributes.
func parseChunkHeader(line string) (name string, attrs map[string]string) {
	attrs = make(map[string]string)

	// Extract name: between << and >>=
	start := strings.Index(line, "<<")
	if start == -1 {
		return "", attrs
	}
	end := strings.Index(line, ">>=")
	if end == -1 {
		return "", attrs
	}
	name = strings.TrimSpace(line[start+2 : end])

	// Parse attributes after >>=.
	rest := strings.TrimSpace(line[end+3:])
	for rest != "" {
		rest = strings.TrimSpace(rest)
		// Colon present → key: value pair
		colonIdx := findUnquotedColon(rest)
		if colonIdx != -1 {
			key := strings.TrimSpace(rest[:colonIdx])
			rest = strings.TrimSpace(rest[colonIdx+1:])

			var value string
			if strings.HasPrefix(rest, "\"") {
				endQuote := strings.Index(rest[1:], "\"")
				if endQuote == -1 {
					value = rest[1:]
					rest = ""
				} else {
					value = rest[1 : endQuote+1]
					rest = strings.TrimSpace(rest[endQuote+2:])
				}
			} else {
				nextAttr := findNextAttrStart(rest)
				if nextAttr == -1 {
					value = strings.TrimSpace(rest)
					rest = ""
				} else {
					value = strings.TrimSpace(rest[:nextAttr])
					rest = strings.TrimSpace(rest[nextAttr:])
				}
			}
			attrs[key] = value
			continue
		}

		// No colon → bare keyword (boolean flag).
		// Tokenize by whitespace.
		token := strings.Fields(rest)[0]
		if token == "override" {
			attrs["override"] = "true"
		}
		rest = ""
	}

	return name, attrs
}

// findUnquotedColon finds a colon that's not inside quotes.
func findUnquotedColon(s string) int {
	inQuote := false
	for i := 0; i < len(s); i++ {
		if s[i] == '"' {
			inQuote = !inQuote
		}
		if !inQuote && s[i] == ':' {
			return i
		}
	}
	return -1
}

// findNextAttrStart finds the start of the next "key:" attribute.
func findNextAttrStart(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' || s[i] == '\t' {
			// Check if after this whitespace there's alphanumeric + :.
			j := i + 1
			for j < len(s) && (s[j] == ' ' || s[j] == '\t') {
				j++
			}
			if j < len(s) {
				// Look for : within a reasonable distance.
				for k := j; k < len(s) && k < j+20; k++ {
					if s[k] == ':' {
						return j
					}
					if s[k] == ' ' || s[k] == '\t' {
						break
					}
				}
			}
		}
	}
	return -1
}

// FormatChunkHeader formats a chunk definition header line.
func FormatChunkHeader(name string, file string, pipeCmd string) string {
	var b strings.Builder
	b.WriteString("<<")
	b.WriteString(name)
	b.WriteString(">>=")
	if file != "" {
		b.WriteString(" file: ")
		b.WriteString(file)
	}
	if pipeCmd != "" {
		b.WriteString(" pipe: ")
		b.WriteString(pipeCmd)
	}
	return b.String()
}

// LanguageExtensions maps file extensions to language names for auto-detection.
// This map is configurable; users can modify it at init time or via
// goweb configuration.
var LanguageExtensions = map[string]string{
	".go":     "go",
	".py":     "python",
	".js":     "javascript",
	".ts":     "typescript",
	".rs":     "rust",
	".c":      "c",
	".h":      "c",
	".cpp":    "cpp",
	".cc":     "cpp",
	".hpp":    "cpp",
	".java":   "java",
	".rb":     "ruby",
	".sh":     "bash",
	".bash":   "bash",
	".zsh":    "bash",
	".pl":     "perl",
	".php":    "php",
	".swift":  "swift",
	".kt":     "kotlin",
	".scala":  "scala",
	".zig":    "zig",
	".md":     "markdown",
	".html":   "html",
	".htm":    "html",
	".css":    "css",
	".scss":   "scss",
	".less":   "less",
	".json":   "json",
	".yaml":   "yaml",
	".yml":    "yaml",
	".toml":   "toml",
	".xml":    "xml",
	".sql":    "sql",
	".r":      "r",
	".lua":    "lua",
	".dart":   "dart",
	".ex":     "elixir",
	".exs":    "elixir",
	".erl":    "erlang",
	".hs":     "haskell",
	".nim":    "nim",
	".vue":    "vue",
	".svelte": "svelte",
}

// DetectLanguage infers the programming language from a file path by
// looking up its extension in LanguageExtensions. Returns "text" if
// no match is found.
func DetectLanguage(filePath string) string {
	ext := ""
	for i := len(filePath) - 1; i >= 0; i-- {
		if filePath[i] == '.' {
			ext = filePath[i:]
			break
		}
		if filePath[i] == '\\' || filePath[i] == '/' {
			break
		}
	}
	if ext != "" {
		if lang, ok := LanguageExtensions[ext]; ok {
			return lang
		}
	}
	return "text"
}

// inferLanguages fills in empty Language fields on all chunks by checking
// their File attribute for a recognizable extension.
func inferLanguages(chunks []*Chunk) {
	for _, c := range chunks {
		if c.Language == "" && c.File != "" {
			c.Language = DetectLanguage(c.File)
		}
	}
}

// ValidateDocument checks for common issues in a parsed document:
// unresolved references, circular dependencies, etc.
// It returns a list of warnings and errors.
func ValidateDocument(doc *Document) error {
	chunkTable, err := doc.ChunkTable()
	if err != nil {
		return err
	}

	refs := doc.AllReferences()
	var missing []string
	for ref := range refs {
		if _, ok := chunkTable[ref]; !ok {
			missing = append(missing, ref)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("unresolved references: %s", strings.Join(missing, ", "))
	}

	return nil
}
