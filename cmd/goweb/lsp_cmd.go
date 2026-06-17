package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/manic/goweb/pkg/parser"
	"github.com/manic/goweb/pkg/preproc"
)

// LSP types
type jsonrpcMessage struct {
	Jsonrpc string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *lspError       `json:"error,omitempty"`
}

type lspError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type initializeParams struct {
	ProcessID    int `json:"processId"`
	Capabilities struct {
		TextDocument struct {
			Completion     map[string]interface{} `json:"completion,omitempty"`
			Definition     map[string]interface{} `json:"definition,omitempty"`
			DocumentSymbol map[string]interface{} `json:"documentSymbol,omitempty"`
		} `json:"textDocument"`
	} `json:"capabilities"`
}

type textDocumentItem struct {
	URI        string `json:"uri"`
	LanguageID string `json:"languageId"`
	Version    int    `json:"version"`
	Text       string `json:"text"`
}

type didOpenParams struct {
	TextDocument textDocumentItem `json:"textDocument"`
}

type didChangeParams struct {
	TextDocument struct {
		URI     string `json:"uri"`
		Version int    `json:"version"`
	} `json:"textDocument"`
	ContentChanges []struct {
		Text string `json:"text"`
	} `json:"contentChanges"`
}

type cursorParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
	Position struct {
		Line      int `json:"line"`
		Character int `json:"character"`
	} `json:"position"`
}

type completionItem struct {
	Label      string `json:"label"`
	Kind       int    `json:"kind"`
	Detail     string `json:"detail,omitempty"`
	InsertText string `json:"insertText,omitempty"`
}

type diagnostic struct {
	Range   lspRange `json:"range"`
	Message string   `json:"message"`
	Source  string   `json:"source"`
}

type lspRange struct {
	Start lspPosition `json:"start"`
	End   lspPosition `json:"end"`
}

type lspPosition struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type publishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []diagnostic `json:"diagnostics"`
}

type location struct {
	URI   string   `json:"uri"`
	Range lspRange `json:"range"`
}

// lspServer holds state for the language server.
type lspServer struct {
	documents  map[string]string          // URI → text content
	chunkCache map[string][]*parser.Chunk // URI → parsed chunks
}

func newLSPServer() *lspServer {
	return &lspServer{
		documents:  make(map[string]string),
		chunkCache: make(map[string][]*parser.Chunk),
	}
}

func registerLspCmd() {
	lspCmd := &cobra.Command{
		Use:   "lsp",
		Short: "Start the goweb Language Server Protocol server",
		Long: `Starts an LSP server on stdin/stdout for editor integration.
Supports:
  - Diagnostics: unresolved references, syntax errors
  - Go-to-definition: jump from <<ref>> to chunk definition
  - Completions: suggest chunk names

Compatible with VS Code, Neovim, Emacs, and any LSP client.

VS Code integration (add to settings.json):
  "lsp": [{
    "command": ["goweb", "lsp"],
    "language": "markdown"
  }]`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			server := newLSPServer()
			return server.run()
		},
	}
	rootCmd.AddCommand(lspCmd)
}

func (s *lspServer) run() error {
	reader := bufio.NewReader(os.Stdin)
	writer := os.Stdout

	for {
		msg, err := readMessage(reader)
		if err != nil {
			return fmt.Errorf("reading message: %w", err)
		}
		response := s.handleMessage(msg)
		if response != nil {
			if err := writeMessage(writer, response); err != nil {
				return fmt.Errorf("writing response: %w", err)
			}
		}
	}
}

func (s *lspServer) handleMessage(msg jsonrpcMessage) *jsonrpcMessage {
	switch msg.Method {
	case "initialize":
		return s.handleInitialize(msg)
	case "textDocument/didOpen":
		s.handleDidOpen(msg)
	case "textDocument/didChange":
		s.handleDidChange(msg)
	case "textDocument/completion":
		return s.handleCompletion(msg)
	case "textDocument/definition":
		return s.handleDefinition(msg)
	case "textDocument/documentSymbol":
		return s.handleDocumentSymbol(msg)
	case "shutdown":
		os.Exit(0)
	}
	return nil
}

func (s *lspServer) handleInitialize(msg jsonrpcMessage) *jsonrpcMessage {
	capabilities := map[string]interface{}{
		"capabilities": map[string]interface{}{
			"textDocumentSync": 1, // full sync
			"completionProvider": map[string]interface{}{
				"triggerCharacters": []string{"<"},
			},
			"definitionProvider":     true,
			"documentSymbolProvider": true,
			"diagnosticsProvider":    true,
		},
	}
	return &jsonrpcMessage{
		Jsonrpc: "2.0",
		ID:      msg.ID,
		Result:  toRaw(capabilities),
	}
}

func (s *lspServer) handleDidOpen(msg jsonrpcMessage) {
	var params didOpenParams
	json.Unmarshal(msg.Params, &params)
	uri := params.TextDocument.URI
	s.documents[uri] = params.TextDocument.Text
	s.updateDiagnostics(uri)
}

func (s *lspServer) handleDidChange(msg jsonrpcMessage) {
	var params didChangeParams
	json.Unmarshal(msg.Params, &params)
	uri := params.TextDocument.URI
	if len(params.ContentChanges) > 0 {
		s.documents[uri] = params.ContentChanges[0].Text
	}
	s.updateDiagnostics(uri)
}

func (s *lspServer) updateDiagnostics(uri string) {
	text, ok := s.documents[uri]
	if !ok {
		return
	}

	// Write to temp file for parsing.
	sourcePath := uriToPath(uri)
	os.MkdirAll(filepath.Dir(sourcePath), 0755)
	os.WriteFile(sourcePath, []byte(text), 0644)

	var diags []diagnostic

	// Check for unresolved references.
	procResult, err := preproc.Process(sourcePath, nil)
	if err != nil {
		diags = append(diags, diagnostic{
			Range:   lspRange{Start: lspPosition{Line: 0, Character: 0}, End: lspPosition{Line: 0, Character: 1}},
			Message: fmt.Sprintf("Preprocessing error: %v", err),
			Source:  "goweb",
		})
	} else {
		doc, err := parser.ParseLines(procResult.Lines, sourcePath)
		if err != nil {
			diags = append(diags, diagnostic{
				Range:   lspRange{Start: lspPosition{Line: 0, Character: 0}, End: lspPosition{Line: 0, Character: 1}},
				Message: fmt.Sprintf("Parse error: %v", err),
				Source:  "goweb",
			})
		} else {
			// Check for unresolved references.
			chunkTable, _ := doc.ChunkTable()
			refs := doc.AllReferences()
			for ref := range refs {
				if _, ok := chunkTable[ref]; !ok {
					// Find the line where this ref appears.
					lineNum := findRefLine(text, ref)
					diags = append(diags, diagnostic{
						Range:   lspRange{Start: lspPosition{Line: lineNum, Character: 0}, End: lspPosition{Line: lineNum, Character: len(ref) + 4}},
						Message: fmt.Sprintf("Unresolved reference: <%s>", ref),
						Source:  "goweb",
					})
				}
			}
		}
		// Cache chunks.
		s.chunkCache[uri] = doc.Chunks
	}

	// Publish diagnostics.
	publishMsg := jsonrpcMessage{
		Jsonrpc: "2.0",
		Method:  "textDocument/publishDiagnostics",
		Params: toRaw(publishDiagnosticsParams{
			URI:         uri,
			Diagnostics: diags,
		}),
	}
	writeMessage(os.Stdout, &publishMsg)
}

func (s *lspServer) handleCompletion(msg jsonrpcMessage) *jsonrpcMessage {
	uri := ""
	var params cursorParams
	json.Unmarshal(msg.Params, &params)
	uri = params.TextDocument.URI

	chunks := s.chunkCache[uri]
	var items []completionItem
	seen := make(map[string]bool)
	for _, c := range chunks {
		if !seen[c.Name] {
			items = append(items, completionItem{
				Label:      c.Name,
				Kind:       3, // Kind field
				Detail:     fmt.Sprintf("chunk (%s: %d)", c.Source, c.Line),
				InsertText: c.Name + ">>",
			})
			seen[c.Name] = true
		}
	}

	return &jsonrpcMessage{
		Jsonrpc: "2.0",
		ID:      msg.ID,
		Result: toRaw(map[string]interface{}{
			"isIncomplete": false,
			"items":        items,
		}),
	}
}

func (s *lspServer) handleDefinition(msg jsonrpcMessage) *jsonrpcMessage {
	var params cursorParams
	json.Unmarshal(msg.Params, &params)
	uri := params.TextDocument.URI
	line := params.Position.Line
	_ = line

	// Find the chunk name at the cursor position.
	text := s.documents[uri]
	if text == "" {
		return nil
	}

	// Scan for <<name>> at or near the cursor line.
	lines := strings.Split(text, "\n")
	if line < 0 || line >= len(lines) {
		return nil
	}
	cursorLine := lines[line]
	col := params.Position.Character

	// Find <<name>> that contains the cursor.
	refName := findRefAtPos(cursorLine, col)
	if refName == "" {
		return nil
	}

	// Find the chunk definition.
	chunks := s.chunkCache[uri]
	for _, c := range chunks {
		if c.Name == refName {
			defLine := c.Line - 1
			return &jsonrpcMessage{
				Jsonrpc: "2.0",
				ID:      msg.ID,
				Result: toRaw(location{
					URI:   uri,
					Range: lspRange{Start: lspPosition{Line: defLine, Character: 0}, End: lspPosition{Line: defLine, Character: 1}},
				}),
			}
		}
	}
	return nil
}

func (s *lspServer) handleDocumentSymbol(msg jsonrpcMessage) *jsonrpcMessage {
	uri := ""
	var params struct {
		TextDocument struct {
			URI string `json:"uri"`
		} `json:"textDocument"`
	}
	json.Unmarshal(msg.Params, &params)
	uri = params.TextDocument.URI

	chunks := s.chunkCache[uri]
	var symbols []map[string]interface{}
	for _, c := range chunks {
		symbols = append(symbols, map[string]interface{}{
			"name": c.Name,
			"kind": 13, // Kind.Function
			"range": lspRange{
				Start: lspPosition{Line: c.Line - 1, Character: 0},
				End:   lspPosition{Line: c.Line, Character: 0},
			},
			"selectionRange": lspRange{
				Start: lspPosition{Line: c.Line - 1, Character: 0},
				End:   lspPosition{Line: c.Line - 1, Character: len(c.Name) + 4},
			},
		})
	}

	return &jsonrpcMessage{
		Jsonrpc: "2.0",
		ID:      msg.ID,
		Result:  toRaw(symbols),
	}
}

// --- Helpers ---

func readMessage(reader *bufio.Reader) (jsonrpcMessage, error) {
	// Read Content-Length header.
	header, err := reader.ReadString('\n')
	if err != nil {
		return jsonrpcMessage{}, err
	}
	header = strings.TrimSpace(header)

	var length int
	if _, err := fmt.Sscanf(header, "Content-Length: %d", &length); err != nil {
		return jsonrpcMessage{}, fmt.Errorf("invalid header: %s", header)
	}

	// Read the blank line.
	reader.ReadString('\n')

	// Read content.
	content := make([]byte, length)
	_, err = reader.Read(content)
	if err != nil {
		return jsonrpcMessage{}, err
	}

	var msg jsonrpcMessage
	if err := json.Unmarshal(content, &msg); err != nil {
		return jsonrpcMessage{}, err
	}
	return msg, nil
}

func writeMessage(writer *os.File, msg *jsonrpcMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(writer, "Content-Length: %d\r\n\r\n%s", len(data), data)
	return err
}

func toRaw(v interface{}) json.RawMessage {
	data, _ := json.Marshal(v)
	return data
}

func uriToPath(uri string) string {
	// file:///c:/path → c:/path
	path := strings.TrimPrefix(uri, "file://")
	return path
}

func findRefLine(text, refName string) int {
	for i, line := range strings.Split(text, "\n") {
		if strings.Contains(line, "<<"+refName+">>") {
			return i
		}
	}
	return 0
}

func findRefAtPos(line string, col int) string {
	// Scan for <<...>> patterns.
	for i := 0; i < len(line); i++ {
		if i+1 < len(line) && line[i] == '<' && line[i+1] == '<' {
			end := strings.Index(line[i+2:], ">>")
			if end == -1 {
				continue
			}
			start := i
			endPos := i + 2 + end + 1
			if col >= start && col <= endPos {
				name := strings.TrimSpace(line[i+2 : i+2+end])
				// Skip <<name>>= definitions.
				if endPos+1 < len(line) && line[endPos+1] != '=' {
					return name
				}
			}
		}
	}
	return ""
}
