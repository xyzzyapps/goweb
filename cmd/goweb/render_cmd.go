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
		tmpl, err := template.New("page").Funcs(templateFuncs()).Parse(string(tmplContent))
		if err != nil {
			return fmt.Errorf("parsing template: %w", err)
		}
		var out bytes.Buffer
		if err := tmpl.Execute(&out, pageData); err != nil {
			return fmt.Errorf("executing template: %w", err)
		}
		finalHTML = out.String()
	} else {
		// Default template: mdBook-inspired documentation theme.
		finalHTML = renderDefaultTheme(pageData)
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

// templateFuncs returns common template functions.
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
	}
}

// renderDefaultTheme generates an mdBook-inspired HTML page with CSS styling.
func renderDefaultTheme(data PageData) string {
	const defaultTmpl = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}}</title>
<style>
  :root {
    --bg: #fff; --text: #24292f; --link: #0969da;
    --code-bg: #f6f8fa; --border: #d0d7de;
    --sidebar-bg: #f6f8fa; --sidebar-width: 260px;
  }
  @media (prefers-color-scheme: dark) {
    :root {
      --bg: #0d1117; --text: #c9d1d9; --link: #58a6ff;
      --code-bg: #161b22; --border: #30363d;
      --sidebar-bg: #161b22;
    }
  }
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif;
    background: var(--bg); color: var(--text); line-height: 1.6;
  }
  .page { max-width: 860px; margin: 0 auto; padding: 40px 24px; }
  h1 { font-size: 2em; border-bottom: 1px solid var(--border); padding-bottom: .3em; margin: .67em 0; }
  h2 { font-size: 1.5em; border-bottom: 1px solid var(--border); padding-bottom: .3em; margin: .83em 0; }
  h3 { font-size: 1.25em; margin: 1em 0; }
  h4 { font-size: 1em; margin: 1.33em 0; }
  p { margin: 1em 0; }
  a { color: var(--link); text-decoration: none; }
  a:hover { text-decoration: underline; }
  code {
    font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
    background: var(--code-bg); padding: .2em .4em; border-radius: 4px; font-size: .85em;
  }
  pre {
    background: var(--code-bg); border: 1px solid var(--border); border-radius: 6px;
    padding: 16px; overflow-x: auto; margin: 1em 0;
  }
  pre code { background: none; padding: 0; border-radius: 0; }
  blockquote {
    border-left: 4px solid var(--border); padding: 0 1em; color: #656d76; margin: 1em 0;
  }
  table { border-collapse: collapse; width: 100%; margin: 1em 0; }
  th, td { border: 1px solid var(--border); padding: 8px 12px; text-align: left; }
  th { background: var(--code-bg); font-weight: 600; }
  ul, ol { padding-left: 2em; margin: 1em 0; }
  hr { border: none; border-top: 1px solid var(--border); margin: 2em 0; }
  .chunk-table { margin: 2em 0; }
  .chunk-table summary { cursor: pointer; font-weight: 600; color: var(--link); }
  .footer {
    margin-top: 3em; padding-top: 1em; border-top: 1px solid var(--border);
    font-size: .85em; color: #656d76;
  }
</style>
</head>
<body>
<div class="page">
<h1>{{.Title}}</h1>
{{.Content}}
{{if .Chunks}}
<details class="chunk-table">
<summary>Code chunks ({{len .Chunks}})</summary>
<table>
<tr><th>Name</th><th>File</th><th>Tags</th><th>Line</th></tr>
{{range .Chunks}}
<tr><td><code>&lt;&lt;{{.Name}}&gt;&gt;</code></td><td>{{.File}}</td><td>{{join .Tags ", "}}</td><td>{{.Line}}</td></tr>
{{end}}
</table>
</details>
{{end}}
<div class="footer">Generated by <a href="https://github.com/manic/goweb">goweb</a></div>
</div>
</body>
</html>`
	tmpl := template.Must(template.New("page").Funcs(template.FuncMap{
		"join": strings.Join,
	}).Parse(defaultTmpl))
	var out bytes.Buffer
	tmpl.Execute(&out, data)
	return out.String()
}
