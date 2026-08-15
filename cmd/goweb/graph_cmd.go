package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/xyzzyapps/goweb/pkg/parser"
	"github.com/xyzzyapps/goweb/pkg/preproc"
)

// registerGraphCmd adds the graph (DOT) subcommand to rootCmd.
func registerGraphCmd() {
	graphCmd := &cobra.Command{
		Use:   "graph <source.md>",
		Short: "Generate a Graphviz DOT dependency diagram",
		Long: `Graph reads a goweb markdown file and outputs a DOT-format
dependency graph showing how chunks reference each other.

Pipe the output through Graphviz to render an image:

  goweb graph program.md | dot -Tpng -o deps.png
  goweb graph program.md | dot -Tsvg -o deps.svg

Use --cluster to group chunks by their output file:

  goweb graph --cluster program.md | dot -Tpng -o deps.png`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			vars := parseVars(varFlags)
			cluster, _ := cmd.Flags().GetBool("cluster")
			return runGraph(args[0], vars, cluster)
		},
	}

	graphCmd.Flags().StringArrayVarP(&varFlags, "var", "v", nil,
		"Set a variable (key=value, can be specified multiple times)")
	graphCmd.Flags().Bool("cluster", false, "Group chunks by their output file")
	rootCmd.AddCommand(graphCmd)
}

// runGraph generates a DOT dependency graph.
func runGraph(sourcePath string, vars map[string]string, cluster bool) error {
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

	// Build dependency map: chunk → chunks it references.
	deps := make(map[string][]string)
	chunkMap := make(map[string]*parser.Chunk)
	for _, c := range doc.Chunks {
		chunkMap[c.Name] = c
		if _, ok := deps[c.Name]; !ok {
			deps[c.Name] = nil
		}
	}

	for _, c := range doc.Chunks {
		refs := extractRefs(c.Body)
		deps[c.Name] = append(deps[c.Name], refs...)
	}

	// Collect chunk names sorted.
	var names []string
	for name := range deps {
		names = append(names, name)
	}
	sort.Strings(names)

	// Build file groups for clustering.
	fileGroups := make(map[string][]string)
	for name := range deps {
		file := ""
		if c, ok := chunkMap[name]; ok {
			file = c.File
		}
		fileGroups[file] = append(fileGroups[file], name)
	}

	// Output DOT header.
	fmt.Println("digraph goweb {")
	fmt.Println("  rankdir=LR;")
	fmt.Println("  node [shape=box, style=rounded];")

	if cluster {
		// Output file-based subgraphs.
		var files []string
		for f := range fileGroups {
			files = append(files, f)
		}
		sort.Strings(files)
		for _, f := range files {
			label := f
			if label == "" {
				label = "no file"
			}
			fmt.Printf("  subgraph cluster_%s {\n", sanitizeID(label))
			fmt.Printf("    label=%q;\n", label)
			fmt.Printf("    style=filled; color=lightgrey;\n")
			for _, name := range fileGroups[f] {
				fmt.Printf("    %s;\n", sanitizeID(name))
			}
			fmt.Println("  }")
		}
	}

	// Output edges.
	for _, name := range names {
		for _, dep := range deps[name] {
			// Only draw edges to chunks that exist.
			if _, ok := chunkMap[dep]; ok {
				fmt.Printf("  %s -> %s;\n", sanitizeID(name), sanitizeID(dep))
			}
		}
	}

	fmt.Println("}")
	return nil
}

// sanitizeID makes a string safe for use as a DOT identifier.
func sanitizeID(s string) string {
	// DOT identifiers can't contain special characters.
	// Quote them if needed.
	if strings.ContainsAny(s, " -./\\<>") {
		return fmt.Sprintf("%q", s)
	}
	return s
}
