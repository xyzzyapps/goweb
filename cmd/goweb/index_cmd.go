package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xyzzyapps/goweb/pkg/parser"
	"github.com/xyzzyapps/goweb/pkg/preproc"
)

// registerIndexCmd adds the index subcommand to rootCmd.
func registerIndexCmd() {
	indexCmd := &cobra.Command{
		Use:   "index <source.md>",
		Short: "Print a cross-reference index of all chunks",
		Long: `Index reads a goweb markdown file and prints a table showing
every chunk, where it is defined, and where it is referenced.

Example:
  goweb index program.md
  goweb index --var debug=true program.md`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			vars := parseVars(varFlags)
			return runIndex(args[0], vars)
		},
	}

	indexCmd.Flags().StringArrayVarP(&varFlags, "var", "v", nil,
		"Set a variable (key=value, can be specified multiple times)")
	rootCmd.AddCommand(indexCmd)
}

// runIndex generates a cross-reference index for a source file.
func runIndex(sourcePath string, vars map[string]string) error {
	// Preprocess.
	preprocResult, err := preproc.Process(sourcePath, vars)
	if err != nil {
		return fmt.Errorf("preprocessing: %w", err)
	}

	// Parse into chunks.
	doc, err := parser.ParseLines(preprocResult.Lines, sourcePath)
	if err != nil {
		return fmt.Errorf("parsing: %w", err)
	}

	// Build reference map: chunk → list of chunks that reference it.
	refs := make(map[string][]string) // chunk name → referenced-by list
	defs := make(map[string]*parser.Chunk)
	for _, c := range doc.Chunks {
		defs[c.Name] = c
		if _, ok := refs[c.Name]; !ok {
			refs[c.Name] = nil
		}
	}

	// Collect all references.
	for _, c := range doc.Chunks {
		extracted := extractRefs(c.Body)
		for _, ref := range extracted {
			refs[ref] = append(refs[ref], c.Name)
		}
	}

	// Print header.
	fmt.Println("=== Cross-Reference Index ===")
	fmt.Println()

	// Collect and sort chunk names.
	var names []string
	for name := range defs {
		names = append(names, name)
	}
	sort.Strings(names)

	// Print per-chunk references.
	for _, name := range names {
		c := defs[name]
		loc := fmt.Sprintf("%s:%d", c.Source, c.Line)
		if c.File != "" {
			fmt.Printf("  <<%s>>  defined at %s  → %s\n", name, loc, c.File)
		} else {
			fmt.Printf("  <<%s>>  defined at %s\n", name, loc)
		}

		// Show references (where this chunk is referenced from).
		if referrers, ok := refs[name]; ok && len(referrers) > 0 {
			// Deduplicate and sort.
			seen := make(map[string]bool)
			var unique []string
			for _, r := range referrers {
				if !seen[r] {
					seen[r] = true
					unique = append(unique, r)
				}
			}
			sort.Strings(unique)
			fmt.Printf("         referenced by: %s\n", strings.Join(unique, ", "))
		} else {
			fmt.Println("         (not referenced)")
		}
		fmt.Println()
	}

	// Print orphaned references (references to undefined chunks).
	var undefined []string
	for ref := range refs {
		if _, ok := defs[ref]; !ok {
			undefined = append(undefined, ref)
		}
	}
	if len(undefined) > 0 {
		sort.Strings(undefined)
		fmt.Println("  !! UNDEFINED REFERENCES:")
		for _, u := range undefined {
			fmt.Printf("     <<%s>> referenced from: %s\n", u, strings.Join(refs[u], ", "))
		}
		fmt.Println()
	}

	// Summary.
	fmt.Printf("  %d chunks defined, %d unique chunk references\n", len(defs), len(refs))
	return nil
}

// extractRefs extracts <<name>> references from a body string, excluding definitions.
func extractRefs(body string) []string {
	var refs []string
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
		// Skip <<name>>= definitions.
		if end+2 < len(body) && body[end+2] == '=' {
			i = end + 3
			continue
		}
		name := strings.TrimSpace(body[start+2 : end])
		if name != "" {
			refs = append(refs, name)
		}
		i = end + 2
	}
	return refs
}
