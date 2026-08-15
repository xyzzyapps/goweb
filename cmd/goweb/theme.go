package main

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
)

// renderDefaultTheme generates a light documentation page inspired by
// GitBook / lit-lang/lit docs: white canvas, light sidebar, blue accents.
func renderDefaultTheme(data PageData) string {
	const defaultTmpl = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}} — goweb</title>
<meta name="description" content="{{.Title}} — literate programming with goweb: noweb-style chunks on GitHub Flavored Markdown. Tangle, weave, and render documentation.">
<meta name="robots" content="index, follow">
<meta name="author" content="goweb contributors (Grok, DeepSeek, Flask, Gemini, Grok 4.6)">
<link rel="canonical" href="https://xyzzyapps.github.io/goweb/">
<meta property="og:type" content="article">
<meta property="og:title" content="{{.Title}} — goweb">
<meta property="og:description" content="Literate programming with noweb-style chunks on GitHub Flavored Markdown.">
<meta property="og:url" content="https://xyzzyapps.github.io/goweb/">
<meta property="og:site_name" content="goweb">
<meta name="twitter:card" content="summary">
<meta name="twitter:title" content="{{.Title}} — goweb">
<meta name="twitter:description" content="Literate programming with noweb-style chunks on GitHub Flavored Markdown.">
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=Source+Code+Pro:wght@400;600&display=swap" rel="stylesheet">
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.11.0/styles/github.min.css">
<script type="application/ld+json">
{
  "@context": "https://schema.org",
  "@graph": [
    {
      "@type": "SoftwareApplication",
      "name": "goweb",
      "applicationCategory": "DeveloperApplication",
      "operatingSystem": "Windows, macOS, Linux",
      "description": "Noweb-style literate programming for GitHub Flavored Markdown.",
      "url": "https://github.com/manic/goweb",
      "author": [
        {"@type": "Organization", "name": "Grok"},
        {"@type": "Organization", "name": "DeepSeek"},
        {"@type": "Organization", "name": "Flask"},
        {"@type": "Organization", "name": "Gemini"},
        {"@type": "SoftwareApplication", "name": "Grok 4.6", "publisher": {"@type": "Organization", "name": "xAI"}}
      ],
      "license": "https://opensource.org/licenses/MIT"
    },
    {
      "@type": "TechArticle",
      "headline": {{printf "%q" .Title}},
      "description": "goweb literate program rendered as HTML documentation.",
      "author": {"@type": "Organization", "name": "goweb"},
      "mainEntityOfPage": "https://xyzzyapps.github.io/goweb/"
    }
  ]
}
</script>
<style>
:root {
  --bg: #ffffff;
  --bg-sidebar: #f7f8fa;
  --bg-code: #f6f8fa;
  --border: #e6e8eb;
  --text: #1c1e21;
  --muted: #5b6169;
  --link: #346ddb;
  --link-hover: #1d4ed8;
  --accent: #346ddb;
  --sidebar-w: 280px;
}
* { box-sizing: border-box; margin: 0; padding: 0; }
html { font-size: 16px; }
body {
  font-family: Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  color: var(--text);
  background: var(--bg);
  line-height: 1.65;
}
.layout { display: flex; min-height: 100vh; }
.sidebar {
  position: fixed; top: 0; bottom: 0; left: 0;
  width: var(--sidebar-w);
  background: var(--bg-sidebar);
  border-right: 1px solid var(--border);
  overflow-y: auto;
  padding: 1.5rem 1.25rem 2.5rem;
}
.brand {
  font-weight: 700; font-size: 1.15rem; color: var(--text);
  text-decoration: none; display: block; margin-bottom: 0.25rem;
}
.brand-sub { font-size: 0.8rem; color: var(--muted); margin-bottom: 1.5rem; }
.nav-label {
  font-size: 0.72rem; font-weight: 600; letter-spacing: 0.06em;
  text-transform: uppercase; color: var(--muted); margin: 1.25rem 0 0.5rem;
}
.sidebar a {
  display: block; color: var(--text); text-decoration: none;
  font-size: 0.92rem; padding: 0.28rem 0.5rem; border-radius: 6px;
}
.sidebar a:hover { background: #eceff3; color: var(--link); }
.sidebar a.current { background: #e8eefc; color: var(--link); font-weight: 600; }
.toc-l1 { padding-left: 0; }
.toc-l2 { padding-left: 0.75rem; }
.toc-l3 { padding-left: 1.5rem; font-size: 0.86rem; color: var(--muted); }
.main { margin-left: var(--sidebar-w); flex: 1; }
.topbar {
  display: none; padding: 0.75rem 1rem; border-bottom: 1px solid var(--border);
  background: var(--bg);
}
.content { max-width: 800px; margin: 0 auto; padding: 2.5rem 2rem 4rem; }
.crumbs { font-size: 0.85rem; color: var(--muted); margin-bottom: 1.5rem; }
.crumbs a { color: var(--link); text-decoration: none; }
article h1, article h2, article h3, article h4 {
  font-weight: 700; line-height: 1.3; margin: 1.6em 0 0.6em;
}
article h1 { font-size: 2rem; margin-top: 0; }
article h2 { font-size: 1.45rem; padding-bottom: 0.3rem; border-bottom: 1px solid var(--border); }
article h3 { font-size: 1.15rem; }
article p { margin: 0 0 1rem; }
article a { color: var(--link); text-decoration: none; }
article a:hover { color: var(--link-hover); text-decoration: underline; }
article code {
  font-family: "Source Code Pro", ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 0.88em; background: var(--bg-code); padding: 0.12em 0.35em;
  border-radius: 4px; border: 1px solid var(--border);
}
article pre {
  font-family: "Source Code Pro", ui-monospace, Menlo, monospace;
  font-size: 0.85rem; line-height: 1.5;
  background: var(--bg-code); border: 1px solid var(--border);
  border-radius: 8px; padding: 1rem 1.1rem; overflow: auto; margin: 0 0 1.25rem;
}
article pre code { background: none; border: 0; padding: 0; }
article blockquote {
  margin: 0 0 1.25rem; padding: 0.75rem 1rem;
  border-left: 3px solid var(--accent); background: #f3f6fd; color: var(--muted);
}
article table { width: 100%; border-collapse: collapse; margin: 0 0 1.25rem; font-size: 0.92rem; }
article th, article td { border: 1px solid var(--border); padding: 0.5rem 0.7rem; text-align: left; }
article th { background: var(--bg-sidebar); font-weight: 600; }
article ul, article ol { margin: 0 0 1.25rem 1.3rem; }
article li { margin-bottom: 0.35rem; }
article hr { border: 0; border-top: 1px solid var(--border); margin: 2rem 0; }
.chunk-table {
  margin: 2rem 0; border: 1px solid var(--border); border-radius: 8px;
  padding: 0.75rem 1rem; background: var(--bg-sidebar);
}
.chunk-table summary { cursor: pointer; font-weight: 600; color: var(--link); }
.chunk-table table { margin-top: 0.75rem; font-size: 0.82rem; }
footer {
  margin-top: 3rem; padding-top: 1.25rem; border-top: 1px solid var(--border);
  font-size: 0.85rem; color: var(--muted);
}
@media (max-width: 800px) {
  .sidebar { transform: translateX(-100%); transition: transform .2s; z-index: 20; }
  .sidebar.open { transform: none; box-shadow: 8px 0 24px rgba(0,0,0,.08); }
  .main { margin-left: 0; }
  .topbar { display: flex; align-items: center; gap: 0.75rem; position: sticky; top: 0; z-index: 10; }
  .menu-btn { border: 1px solid var(--border); background: #fff; border-radius: 6px; padding: 0.35rem 0.6rem; cursor: pointer; }
}
</style>
</head>
<body>
<div class="layout">
  <aside class="sidebar" id="sidebar">
    <a class="brand" href="#">goweb</a>
    <div class="brand-sub">Literate programming</div>
    <div class="nav-label">Page</div>
    <a class="current" href="#">{{.Title}}</a>
    {{if .Headings}}
    <div class="nav-label">On this page</div>
    {{range .Headings}}
    <a class="toc-l{{.Level}}" href="#{{.ID}}">{{.Text}}</a>
    {{end}}
    {{end}}
  </aside>
  <div class="main">
    <div class="topbar">
      <button class="menu-btn" type="button" onclick="document.getElementById('sidebar').classList.toggle('open')" aria-label="Open menu">☰</button>
      <strong>{{.Title}}</strong>
    </div>
    <div class="content">
      <nav class="crumbs" aria-label="Breadcrumb"><a href="#">Home</a> / {{.Title}}</nav>
      <article itemscope itemtype="https://schema.org/TechArticle">
        <meta itemprop="headline" content="{{.Title}}">
        {{.Content}}
      </article>
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
      <footer>
        <p>Written with Grok, DeepSeek, Flask, Gemini, and Grok 4.6. Generated by <a href="https://github.com/manic/goweb">goweb</a>.</p>
      </footer>
    </div>
  </div>
</div>
<script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.11.0/highlight.min.js"></script>
<script>hljs.highlightAll();</script>
</body>
</html>`
	tmpl := template.Must(template.New("page").Funcs(template.FuncMap{
		"join":   strings.Join,
		"printf": fmt.Sprintf,
	}).Parse(defaultTmpl))
	var out bytes.Buffer
	if err := tmpl.Execute(&out, data); err != nil {
		return ""
	}
	return out.String()
}
