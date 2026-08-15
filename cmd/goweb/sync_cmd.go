package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xyzzyapps/goweb/pkg/parser"
	"github.com/xyzzyapps/goweb/pkg/preproc"
	"github.com/xyzzyapps/goweb/pkg/tangle"
)

func registerSyncCmd() {
	syncCmd := &cobra.Command{
		Use:   "sync <source.md>",
		Short: "Sync tangled output back to the source .md file",
		Long: `Sync reads the tangled output files (using #line directives to
map lines back to chunks) and updates the source .md file with any changes
made directly to the generated code.

This enables a basic bidirectional workflow:
  1. goweb tangle --line-directives program.md  (generate code with #line refs)
  2. Edit the generated .go / .py files
  3. goweb sync program.md                       (apply changes back to .md)

Example:
  goweb tangle --line-directives program.md
  # edit main.go
  goweb sync program.md`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			vars := parseVars(varFlags)
			dryRunSync, _ := cmd.Flags().GetBool("dry-run")
			return runSync(args[0], vars, dryRunSync)
		},
	}

	syncCmd.Flags().StringArrayVarP(&varFlags, "var", "v", nil,
		"Set a variable (key=value, can be specified multiple times)")
	syncCmd.Flags().Bool("dry-run", false, "Show what would change without writing")
	rootCmd.AddCommand(syncCmd)
}

// lineDirectiveRe matches //line "file":N and #line N "file"
var lineDirectiveRe = regexp.MustCompile(`//line\s+"([^"]+)":(\d+)|#line\s+(\d+)\s+"([^"]+)"`)

// runSync syncs changes from tangled output files back to the .md source.
func runSync(sourcePath string, vars map[string]string, dryRun bool) error {
	// Read the original source to preserve formatting.
	sourceData, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("reading source: %w", err)
	}
	originalSource := string(sourceData)

	// First, tangle normally to get the current state.
	tangleResult, err := tangleToFile(sourcePath, vars)
	if err != nil {
		return fmt.Errorf("tangling: %w", err)
	}

	// Now parse the source to find chunks.
	preprocResult, err := preproc.Process(sourcePath, vars)
	if err != nil {
		return fmt.Errorf("preprocessing: %w", err)
	}
	doc, err := parser.ParseLines(preprocResult.Lines, sourcePath)
	if err != nil {
		return fmt.Errorf("parsing: %w", err)
	}

	// Build chunk lookup by line number.
	chunksByLine := make(map[int]*parser.Chunk)
	for _, c := range doc.Chunks {
		chunksByLine[c.Line] = c
	}

	// For each tangled output file, parse #line directives and extract chunk bodies.
	updatedBodies := make(map[int]string) // source line → new body
	for _, c := range doc.Chunks {
		if c.File == "" {
			continue
		}
		outPath := filepath.Join(tangleResult.OutputDir, c.File)
		data, err := os.ReadFile(outPath)
		if err != nil {
			continue // File might not exist yet
		}
		lines := strings.Split(string(data), "\n")

		// Find lines belonging to this chunk by scanning for #line directives.
		inChunk := false
		var chunkLines []string
		for _, line := range lines {
			matches := lineDirectiveRe.FindStringSubmatch(line)
			if matches != nil {
				var file string
				var lineNum int
				if matches[1] != "" {
					file = matches[1]
					_, _ = fmt.Sscanf(matches[2], "%d", &lineNum)
				} else {
					_, _ = fmt.Sscanf(matches[3], "%d", &lineNum)
					file = matches[4]
				}
				if filepath.Base(file) == filepath.Base(c.Source) && lineNum == c.Line {
					inChunk = true
					chunkLines = nil
					continue
				} else {
					inChunk = false
				}
			} else if inChunk {
				chunkLines = append(chunkLines, line)
			}
		}

		if len(chunkLines) > 0 {
			newBody := strings.Join(chunkLines, "\n")
			// Remove trailing empty lines.
			newBody = strings.TrimRight(newBody, "\n")
			if newBody != c.Body {
				updatedBodies[c.Line] = newBody
			}
		}
	}

	if len(updatedBodies) == 0 {
		fmt.Println("No changes detected.")
		return nil
	}

	// Update the source .md content.
	sourceLines := strings.Split(originalSource, "\n")
	for lineNum, newBody := range updatedBodies {
		c := chunksByLine[lineNum]
		if c == nil {
			continue
		}
		// Find the chunk body in the source lines and replace it.
		// The chunk starts at line c.Line (1-based) in the source.
		// We need to find the >> terminator.
		startIdx := c.Line - 1 // 0-based slice index
		if startIdx < 0 || startIdx >= len(sourceLines) {
			continue
		}
		// Find the >> line.
		endIdx := -1
		for j := startIdx + 1; j < len(sourceLines); j++ {
			if strings.TrimSpace(sourceLines[j]) == ">>" {
				endIdx = j
				break
			}
		}
		if endIdx == -1 {
			continue
		}

		if dryRun {
			fmt.Printf("=== %s:%d (%s) ===\n", c.Source, c.Line, c.Name)
			fmt.Printf("  old body: %q\n", c.Body)
			fmt.Printf("  new body: %q\n", newBody)
			continue
		}

		// Replace lines between start+1 and endIdx-1 with new body lines.
		var newLines []string
		newLines = append(newLines, sourceLines[:startIdx+1]...)
		newLines = append(newLines, strings.Split(newBody, "\n")...)
		newLines = append(newLines, sourceLines[endIdx:]...)
		sourceLines = newLines
	}

	if dryRun {
		return nil
	}

	// Write the updated source.
	updated := strings.Join(sourceLines, "\n")
	if err := os.WriteFile(sourcePath, []byte(updated), 0644); err != nil {
		return fmt.Errorf("writing source: %w", err)
	}
	fmt.Fprintf(os.Stderr, "updated %s (%d chunks synced)\n", sourcePath, len(updatedBodies))
	return nil
}

// tangleResult holds the output of a tangle operation.
type tangleResult struct {
	OutputDir string
}

// tangleToFile runs a tangle and returns info about the output.
func tangleToFile(sourcePath string, vars map[string]string) (*tangleResult, error) {
	preprocResult, err := preproc.Process(sourcePath, vars)
	if err != nil {
		return nil, err
	}
	doc, err := parser.ParseLines(preprocResult.Lines, sourcePath)
	if err != nil {
		return nil, err
	}
	doc.Vars = vars

	t := tangle.New()
	t.LineDirectives = true // Always use line directives for sync
	// Use a temp output dir or the source file's directory.
	t.OutputDir = filepath.Dir(sourcePath)
	if err := t.Tangle(doc); err != nil {
		return nil, err
	}
	return &tangleResult{OutputDir: t.OutputDir}, nil
}
