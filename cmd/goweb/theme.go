package main

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
)

// renderDefaultTheme generates HTML in the style of https://lit-lang.org/:
// system UI fonts, amber primary, hard offset shadows, centered header + main.
func renderDefaultTheme(data PageData) string {
	const defaultTmpl = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}} — goweb</title>
<meta name="description" content="{{.Title}} — literate programming with goweb: noweb-style chunks on GitHub Flavored Markdown. Tangle, weave, and render documentation.">
<meta name="robots" content="index, follow">
<meta name="author" content="{{if .Author}}{{.Author}}{{else}}DeepSeek and Grok 4.6{{end}}">
<link rel="canonical" href="{{if .Pages}}{{.Pages}}{{else}}#{{end}}">
<meta property="og:type" content="article">
<meta property="og:title" content="{{.Title}} — goweb">
<meta property="og:description" content="Literate programming with noweb-style chunks on GitHub Flavored Markdown.">
<meta property="og:url" content="{{if .Pages}}{{.Pages}}{{else}}#{{end}}">
<meta property="og:site_name" content="goweb">
<meta name="twitter:card" content="summary">
<meta name="twitter:title" content="{{.Title}} — goweb">
<meta name="twitter:description" content="Literate programming with noweb-style chunks on GitHub Flavored Markdown.">
<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.11.0/styles/atom-one-light.min.css">
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
      "url": {{printf "%q" .Repo}},
      "author": [
        {"@type": "Organization", "name": "DeepSeek"},
        {"@type": "SoftwareApplication", "name": "Grok 4.6", "publisher": {"@type": "Organization", "name": "xAI"}}
      ],
      "license": "https://creativecommons.org/licenses/by-sa/4.0/"
    },
    {
      "@type": "TechArticle",
      "headline": {{printf "%q" .Title}},
      "description": "goweb literate program rendered as HTML documentation.",
      "author": {"@type": "Organization", "name": "goweb"},
      "mainEntityOfPage": {{printf "%q" .Pages}}
    }
  ]
}
</script>
<style>
:root {
  --color-primary: #ffcd42;
  --color-primary-light: rgba(255, 205, 66, 0.1);
  --color-primary-darker: #291f04;
  --color-primary-text: #333;
  --gradient-primary: linear-gradient(to right, var(--color-primary), #ff6102);
  --gradient-underline: linear-gradient(transparent, transparent) no-repeat 0 0,
    linear-gradient(transparent, transparent) no-repeat 0 0,
    var(--gradient-primary) no-repeat 0 calc(100% - 0.5px) / 100% 2px;
  --color-code-background: #f8f8f8;
  --color-code-foreground: #383a42;
}
*, *::before, *::after { box-sizing: border-box; }
html { -webkit-text-size-adjust: none; text-size-adjust: none; }
body {
  margin: 0;
  min-height: 100vh;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Noto Sans", Helvetica, Arial, sans-serif, "Apple Color Emoji", "Segoe UI Emoji";
  font-size: 18px;
  line-height: 1.5;
  color: var(--color-primary-text);
  background: #fff;
  accent-color: black;
}
header { padding: 1rem; }
.Logo {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  gap: 0.35rem; width: fit-content; margin: 0 auto; text-decoration: none; color: inherit;
}
.Logo p { font-size: 2rem; font-weight: bold; margin: 0; }
.Nav ul {
  list-style: none; display: flex; gap: 2rem; justify-content: center;
  padding: 0; margin: 1rem 0 0;
}
.Nav a { color: black; font-weight: 600; text-decoration: none; }
.Nav a:hover { color: #000; opacity: 0.65; }
main { max-width: 800px; margin: 0 auto; padding: 1rem 1rem 3rem; }
article h1, article h2, article h3, article h4 { color: black; margin: 2rem 0 1rem; line-height: 1.1; text-wrap: balance; }
article h1, article h2 { padding-bottom: 0.5rem; margin-bottom: 0; }
article h1 { font-size: 2em; }
article p { margin: 1rem 0; }
article a { color: black; font-weight: 600; text-decoration: none; }
article a:hover { opacity: 0.65; }
.chunk-table a { text-decoration: none; }
article code {
  font-family: ui-monospace, SFMono-Regular, Menlo, Consolas, monospace;
  background: var(--color-code-background);
  border: 1px solid var(--color-primary-text);
  border-radius: 0.15rem;
  box-shadow: 1px 1px 0 var(--color-primary-text);
  font-size: 16px;
  padding: 0 0.2rem;
}
article pre {
  white-space: pre; overflow: auto;
  border-radius: 0.5rem; padding: 0.75rem;
  background: var(--color-code-background);
  color: var(--color-code-foreground);
  border: 2px solid var(--color-primary-text);
  box-shadow: 4px 4px 0 var(--color-primary-text);
  font-size: 14px;
}
article pre code { background: none; border: 0; box-shadow: none; padding: 0; font-size: inherit; }
article aside, article blockquote {
  background: var(--color-primary-light);
  border: 2px solid var(--color-primary);
  border-radius: 0.5rem;
  box-shadow: 4px 4px 0 var(--color-primary);
  padding: 1rem; margin: 1rem 0;
}
article table {
  width: 100%; border-collapse: collapse; margin: 1.5rem 0; font-size: 0.9rem;
  box-shadow: 4px 4px 0 var(--color-primary-text);
}
article th { background: var(--color-primary); font-weight: 600; }
article th, article td { padding: 0.75rem 1rem; border: 2px solid var(--color-primary-text); text-align: left; }
article tr:nth-child(even) { background: var(--color-primary-light); }
article ul, article ol { margin: 1rem 0 1rem 1.4rem; }
.chunk-table {
  margin: 2rem 0; padding: 1rem;
  background: var(--color-primary-light);
  border: 2px solid var(--color-primary);
  border-radius: 0.5rem;
  box-shadow: 4px 4px 0 var(--color-primary);
}
.chunk-table summary { cursor: pointer; font-weight: 600; }
footer {
  background: var(--color-primary-darker);
  padding: 2rem 1rem; color: #fff; text-align: center; font-size: 0.9rem;
}
footer a { color: var(--color-primary); font-weight: 600; }
footer h2 { color: var(--color-primary); background: none; margin: 0 0 0.5rem; }
</style>
</head>
<body>
<header>
  <a class="Logo" href="#">
    <p>goweb</p>
  </a>
  <nav class="Nav" aria-label="Main navigation">
    <ul>
      <li><a href="{{if .Repo}}{{.Repo}}{{else}}#{{end}}">Source</a></li>
      <li><a href="#chunks">Documentation</a></li>
    </ul>
  </nav>
</header>
<main>
  <article itemscope itemtype="https://schema.org/TechArticle">
    <meta itemprop="headline" content="{{.Title}}">
    {{.Content}}
  </article>
  {{if .Chunks}}
  <details class="chunk-table" id="chunks">
    <summary>Code chunks ({{len .Chunks}})</summary>
    <table>
      <tr><th>Name</th><th>File</th><th>Tags</th><th>Line</th></tr>
      {{range .Chunks}}
      <tr><td><a href="#{{.ID}}"><code>&lt;&lt;{{.Name}}&gt;&gt;</code></a></td><td>{{.File}}</td><td>{{join .Tags ", "}}</td><td>{{.Line}}</td></tr>
      {{end}}
    </table>
  </details>
  {{end}}
</main>
<footer>
  <h2>goweb</h2>
  <p>Written with DeepSeek and Grok 4.6.{{if .Author}} Copyright {{.Author}}{{if .Email}} &lt;{{.Email}}&gt;{{end}}.{{end}} Generated by <a href="{{if .Repo}}{{.Repo}}{{else}}#{{end}}">goweb</a>.</p>
</footer>
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
