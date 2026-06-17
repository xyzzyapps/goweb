package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"

	"github.com/manic/goweb/pkg/parser"
	"github.com/manic/goweb/pkg/preproc"
	"github.com/manic/goweb/pkg/tangle"
	"github.com/manic/goweb/pkg/weave"
)

var (
	varFlags       []string
	pipeDir        string
	outputDir      string
	dryRun         bool
	lineDirectives bool
	watchMode      bool
	outputFile     string
	rootCmd        *cobra.Command
)

func main() {
	rootCmd = &cobra.Command{
		Use:   "goweb",
		Short: "goweb — noweb-style literate programming with GFM Markdown",
		Long: `goweb is a literate programming tool that uses GFM Markdown with
noweb-style <<chunk>>= syntax. Use 'tangle' to extract code and 'weave' to
generate documentation.`,
	}

	tangleCmd := &cobra.Command{
		Use:   "tangle [flags] <source.md> [chunk]",
		Short: "Extract and resolve code chunks into source files",
		Long: `Tangle reads a goweb markdown file and extracts code chunks,
resolving all <<ref>> references, applying pipe commands, and writing
the output files specified by 'file:' attributes.

If a chunk name is given as the second argument, only that chunk is
resolved and written to stdout.

Examples:
  goweb tangle --var debug=true program.md
  goweb tangle --var debug=false --dry-run program.md
  goweb tangle program.md main > main.go
  goweb tangle --pipe-dir ./scripts program.md`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			vars := parseVars(varFlags)
			sourcePath := args[0]

			if watchMode {
				return watchTangle(sourcePath, vars)
			}

			return runTangle(sourcePath, vars, args)
		},
	}

	initCmd := &cobra.Command{
		Use:   "init [directory]",
		Short: "Scaffold a new literate programming project",
		Long: `Creates a skeleton goweb project with a sample .md file and
optional configuration.

Example:
  goweb init myproject
  cd myproject
  goweb tangle program.md`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := "."
			if len(args) == 1 {
				dir = args[0]
				if err := os.MkdirAll(dir, 0755); err != nil {
					return fmt.Errorf("creating directory: %w", err)
				}
			}
			return runInit(dir)
		},
	}

	weaveCmd := &cobra.Command{
		Use:   "weave [flags] <source.md>",
		Short: "Generate clean markdown documentation",
		Long: `Weave reads a goweb markdown file and strips all goweb control
syntax (<<>> directives, chunk definitions, terminators), producing clean
markdown documentation.

Examples:
  goweb weave program.md > README.md
  goweb weave --var debug=false program.md`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			vars := parseVars(varFlags)
			sourcePath := args[0]

			var w *os.File
			if outputFile != "" {
				f, err := os.Create(outputFile)
				if err != nil {
					return fmt.Errorf("creating output file: %w", err)
				}
				defer f.Close()
				w = f
			} else {
				w = os.Stdout
			}

			return weave.Weave(sourcePath, vars, w, outputDir)
		},
	}

	// Global flags.
	tangleCmd.Flags().StringArrayVarP(&varFlags, "var", "v", nil,
		"Set a variable (key=value, can be specified multiple times)")
	tangleCmd.Flags().StringVarP(&pipeDir, "pipe-dir", "P", "",
		"Working directory for pipe commands")
	tangleCmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false,
		"Print what would be written without actually writing")
	tangleCmd.Flags().BoolVar(&lineDirectives, "line-directives", false,
		"Insert #line///line directives in tangled output pointing to the .md source")
	tangleCmd.Flags().BoolVarP(&watchMode, "watch", "w", false,
		"Watch source file for changes and re-tangle automatically")
	tangleCmd.Flags().StringVarP(&outputDir, "output-dir", "O", "",
		"Base directory for tangled output files")

	weaveCmd.Flags().StringArrayVarP(&varFlags, "var", "v", nil,
		"Set a variable (key=value, can be specified multiple times)")
	weaveCmd.Flags().StringVarP(&outputFile, "output", "o", "",
		"Output file (default: stdout)")
	weaveCmd.Flags().StringVarP(&outputDir, "output-dir", "O", "",
		"Output directory for weave output (filename derived from source)")

	rootCmd.AddCommand(tangleCmd, weaveCmd, initCmd)
	registerIndexCmd()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// runTangle performs the tangle operation on a source file.
func runTangle(sourcePath string, vars map[string]string, args []string) error {
	preprocResult, err := preproc.Process(sourcePath, vars)
	if err != nil {
		return fmt.Errorf("preprocessing: %w", err)
	}
	doc, err := parser.ParseLines(preprocResult.Lines, sourcePath)
	if err != nil {
		return fmt.Errorf("parsing: %w", err)
	}
	doc.Vars = vars

	t := tangle.New()
	t.PipeDir = pipeDir
	t.DryRun = dryRun
	t.LineDirectives = lineDirectives
	t.OutputDir = outputDir

	if len(args) == 2 {
		return t.TangleChunk(doc, args[1], os.Stdout)
	}
	return t.Tangle(doc)
}

// watchTangle watches a source file for changes and re-tangles automatically.
func watchTangle(sourcePath string, vars map[string]string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("creating watcher: %w", err)
	}
	defer watcher.Close()

	absPath, _ := filepath.Abs(sourcePath)
	if err := watcher.Add(absPath); err != nil {
		return fmt.Errorf("watching %s: %w", absPath, err)
	}

	fmt.Fprintf(os.Stderr, "watching %s for changes...\n", absPath)
	// Do an initial tangle.
	if err := runTangle(sourcePath, vars, []string{sourcePath}); err != nil {
		fmt.Fprintf(os.Stderr, "initial tangle error: %v\n", err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
				fmt.Fprintf(os.Stderr, "\nchange detected: %s\n", event.Name)
				time.Sleep(100 * time.Millisecond) // debounce
				if err := runTangle(sourcePath, vars, []string{sourcePath}); err != nil {
					fmt.Fprintf(os.Stderr, "tangle error: %v\n", err)
				} else {
					fmt.Fprintf(os.Stderr, "re-tangled successfully\n")
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(os.Stderr, "watch error: %v\n", err)
		}
	}
}

// runInit scaffolds a new goweb project.
func runInit(dir string) error {
	content := `# My Literate Program

This is a goweb literate programming project.
Write documentation in markdown and embed code in chunks.

<<package>>= file: main.go
package main

import "fmt"

func main() {
	fmt.Println("Hello from goweb!")
}
>>
`
	path := filepath.Join(dir, "program.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	fmt.Fprintf(os.Stderr, "wrote %s\n", path)
	fmt.Fprintf(os.Stderr, "Run: goweb tangle %s\n", path)
	return nil
}
func parseVars(flags []string) map[string]string {
	vars := make(map[string]string)
	for _, f := range flags {
		parts := strings.SplitN(f, "=", 2)
		if len(parts) == 2 {
			vars[parts[0]] = parts[1]
		} else {
			vars[parts[0]] = "true"
		}
	}
	return vars
}
