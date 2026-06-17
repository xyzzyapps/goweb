package main

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"

	"github.com/manic/goweb/pkg/parser"
	"github.com/manic/goweb/pkg/preproc"
	"github.com/manic/goweb/pkg/weave"
)

// PageData is the template context passed to HTML templates.
type PageData struct {
	Title   string
	Content template.HTML
	Source  string
	Chunks  []ChunkInfo
}

// ChunkInfo holds metadata about a chunk for templates.
type ChunkInfo struct {
	Name     string
	Language string
	File     string
	Tags     []string
	Line     int
}

// SitePage holds data for a single page in site generation.
type SitePage struct {
	Title    string
	Content  template.HTML
	Path     string // relative path like "guide/install"
	Source   string // source .md path
	Chunks   []ChunkInfo
	Children []SitePage
}

func registerRenderCmd() {
	renderCmd := &cobra.Command{
		Use:   "render <source.md> [source2.md ...]",
		Short: "Render literate programs to HTML using templates",
		Long: `Render weaves a goweb markdown file and renders it to HTML
using Go templates (similar to Jinja2).

Single file mode:
  goweb render program.md --template page.tmpl > output.html

Site mode (process a directory):
  goweb render --site docs/ --template page.tmpl --output-dir site/

Template variables:
  {{.Title}}    — First # heading from the document
  {{.Content}}  — Rendered HTML content (markdown converted)
  {{.Source}}   — Source .md file path
  {{.Chunks}}   — List of chunks with .Name, .Language, .File, .Tags

Example template (page.tmpl):
  <html><body>
    <h1>{{.Title}}</h1>
    <div>{{.Content}}</div>
  </body></html>`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			vars := parseVars(varFlags)
			siteDir, _ := cmd.Flags().GetString("site")
			tmplPath, _ := cmd.Flags().GetString("template")

			if siteDir != "" {
				return renderSite(siteDir, tmplPath, vars)
			}

			// Single file mode — use first arg.
			outputPath, _ := cmd.Flags().GetString("output")
			return renderFile(args[0], tmplPath, outputPath, vars)
		},
	}

	renderCmd.Flags().StringP("template", "t", "", "Go template file for HTML rendering")
	renderCmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	renderCmd.Flags().StringP("site", "s", "", "Site mode: process a directory of .md files")
	renderCmd.Flags().StringVarP(&outputDir, "output-dir", "O", "", "Output directory for site mode")
	renderCmd.Flags().StringArrayVarP(&varFlags, "var", "v", nil,
		"Set a variable (key=value, can be specified multiple times)")
	rootCmd.AddCommand(renderCmd)
}

// renderFile processes a single .md file and renders it to HTML.
func renderFile(sourcePath, tmplPath, outputPath string, vars map[string]string) error {
	// Weave: strip goweb syntax.
	var buf bytes.Buffer
	if err := weave.Weave(sourcePath, vars, &buf, ""); err != nil {
		return fmt.Errorf("weaving: %w", err)
	}
	markdownContent := buf.String()

	// Convert markdown to HTML.
	mdRenderer := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
	var htmlBuf bytes.Buffer
	if err := mdRenderer.Convert([]byte(markdownContent), &htmlBuf); err != nil {
		return fmt.Errorf("rendering markdown: %w", err)
	}

	// Build page data.
	title := extractTitle(markdownContent)
	chunks := extractChunkInfo(sourcePath, vars)
	pageData := PageData{
		Title:   title,
		Content: template.HTML(htmlBuf.String()),
		Source:  sourcePath,
		Chunks:  chunks,
	}

	// Render through template or output raw HTML.
	var finalHTML string
	if tmplPath != "" {
		tmplContent, err := os.ReadFile(tmplPath)
		if err != nil {
			return fmt.Errorf("reading template: %w", err)
		}
		tmpl, err := template.New("page").Parse(string(tmplContent))
		if err != nil {
			return fmt.Errorf("parsing template: %w", err)
		}
		var out bytes.Buffer
		if err := tmpl.Execute(&out, pageData); err != nil {
			return fmt.Errorf("executing template: %w", err)
		}
		finalHTML = out.String()
	} else {
		// Default template: basic HTML page.
		const defaultTmpl = `<html><head><meta charset="utf-8"><title>{{.Title}}</title></head><body>{{.Content}}</body></html>`
		tmpl, _ := template.New("page").Parse(defaultTmpl)
		var out bytes.Buffer
		tmpl.Execute(&out, pageData)
		finalHTML = out.String()
	}

	// Write output.
	if outputPath != "" {
		if err := os.WriteFile(outputPath, []byte(finalHTML), 0644); err != nil {
			return fmt.Errorf("writing output: %w", err)
		}
		fmt.Fprintf(os.Stderr, "wrote %s\n", outputPath)
	} else {
		fmt.Print(finalHTML)
	}
	return nil
}

// renderSite processes a directory of .md files into a static HTML site.
func renderSite(siteDir, tmplPath string, vars map[string]string) error {
	if tmplPath == "" {
		return fmt.Errorf("--template is required for site mode")
	}
	if outputDir == "" {
		return fmt.Errorf("--output-dir is required for site mode")
	}

	tmplContent, err := os.ReadFile(tmplPath)
	if err != nil {
		return fmt.Errorf("reading template: %w", err)
	}
	tmpl, err := template.New("page").Funcs(template.FuncMap{
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
	}).Parse(string(tmplContent))
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	// Walk the directory for .md files.
	var pages []SitePage
	err = filepath.Walk(siteDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		page, err := buildPage(path, siteDir, vars, tmpl)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error processing %s: %v\n", path, err)
			return nil
		}
		pages = append(pages, page)
		return nil
	})
	if err != nil {
		return fmt.Errorf("walking site directory: %w", err)
	}

	if len(pages) == 0 {
		return fmt.Errorf("no .md files found in %s", siteDir)
	}

	// Write pages.
	for _, page := range pages {
		var out bytes.Buffer
		if err := tmpl.Execute(&out, page); err != nil {
			fmt.Fprintf(os.Stderr, "error rendering %s: %v\n", page.Path, err)
			continue
		}
		outPath := filepath.Join(outputDir, page.Path+".html")
		outDir := filepath.Dir(outPath)
		if err := os.MkdirAll(outDir, 0755); err != nil {
			return fmt.Errorf("creating directory: %w", err)
		}
		if err := os.WriteFile(outPath, out.Bytes(), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", outPath, err)
		}
		fmt.Fprintf(os.Stderr, "wrote %s\n", outPath)
	}

	fmt.Fprintf(os.Stderr, "rendered %d pages\n", len(pages))
	return nil
}

// buildPage processes a single .md file into a SitePage.
func buildPage(path, siteDir string, vars map[string]string, tmpl *template.Template) (SitePage, error) {
	relPath, _ := filepath.Rel(siteDir, path)
	relPath = strings.TrimSuffix(relPath, ".md")

	var buf bytes.Buffer
	if err := weave.Weave(path, vars, &buf, ""); err != nil {
		return SitePage{}, fmt.Errorf("weaving: %w", err)
	}
	markdownContent := buf.String()

	// Convert markdown to HTML.
	mdRenderer := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
	var htmlBuf bytes.Buffer
	if err := mdRenderer.Convert([]byte(markdownContent), &htmlBuf); err != nil {
		return SitePage{}, fmt.Errorf("rendering markdown: %w", err)
	}

	title := extractTitle(markdownContent)
	chunks := extractChunkInfo(path, vars)

	return SitePage{
		Title:   title,
		Content: template.HTML(htmlBuf.String()),
		Path:    relPath,
		Source:  path,
		Chunks:  chunks,
	}, nil
}

// extractTitle gets the first # heading from markdown content.
func extractTitle(markdown string) string {
	for _, line := range strings.Split(markdown, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			return strings.TrimPrefix(trimmed, "# ")
		}
	}
	return filepath.Base(markdown)
}

// extractChunkInfo parses the source and returns chunk metadata.
func extractChunkInfo(sourcePath string, vars map[string]string) []ChunkInfo {
	preprocResult, err := preproc.Process(sourcePath, vars)
	if err != nil {
		return nil
	}
	doc, err := parser.ParseLines(preprocResult.Lines, sourcePath)
	if err != nil {
		return nil
	}
	var info []ChunkInfo
	for _, c := range doc.Chunks {
		info = append(info, ChunkInfo{
			Name:     c.Name,
			Language: c.Language,
			File:     c.File,
			Tags:     c.Tags,
			Line:     c.Line,
		})
	}
	return info
}
