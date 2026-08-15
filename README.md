# goweb

**goweb** is a literate programming tool that brings noweb-style chunk syntax to GitHub Flavored Markdown. Write documentation and code in a single `.md` file, then **tangle** to extract source files, **weave** to generate clean markdown, or **sync** to apply edits back from the generated code.

This codebase was written with **Grok**, **DeepSeek**, **Flask**, and **Gemini**. Later edits (spec, light theme, example, history cleanup, and GitHub Pages) were made by **Grok 4.6** (xAI). See [SPEC.md](SPEC.md) for the full language and architecture contract (for humans and agents).

## Features

- **Noweb-style chunks** — `<<name>>=` defines a chunk, `>>` ends it
- **Out-of-order definitions** — chunks can appear in any order; references are resolved after parsing
- **Cross-file references** — `<<import "file.md">>` loads chunks from other files
- **Chunk overriding** — `<<name>>= override` replaces previous definitions
- **Override imports** — `<<override "file.md">>` loads chunks that replace any existing
- **Conditional blocks** — `<<if var>>` / `<<elif var>>` / `<<else>>` / `<<end>>` controlled by `--var` flags
- **Inline variables** — `{{var}}` substitution inside chunk bodies
- **Code execution** — `exec:` attribute runs chunk body through an interpreter, stdout replaces content
- **Session-based execution** — `session:` attribute shares interpreter state across chunks
- **Pipe commands** — `pipe:` attribute pipes chunk content through external commands (e.g., `gofmt`)
- **File output** — `file:` attribute on a chunk writes to a specific path
- **Chunk metadata** — `tags:` attribute for categorization and filtering
- **`--output-dir`** — base directory for all tangled/weave output files
- **Source line directives** — `--line-directives` inserts `//line`/`#line` comments pointing to the `.md` source
- **Bidirectional sync** — `goweb sync` reads `#line` directives to apply source edits back to the `.md`
- **Watch mode** — `--watch` re-tangles automatically when the source file changes
- **Language auto-detect** — infers fenced-block language from `file:` extension
- **Escape `\>>`** — literal `>>` inside chunk bodies
- **Topological sorting** — chunks are emitted in dependency order
- **Cycle detection** — circular references are caught and reported
- **Weave** — strips goweb syntax to produce clean markdown
- **Cross-reference index** — `goweb index` prints a table of all chunks and their references
- **Dependency diagrams** — `goweb graph` outputs Graphviz DOT format
- **Source → literate** — `goweb reverse` converts source files into `.md` chunks
- **Project scaffolding** — `goweb init` creates a new literate program skeleton
- **YAML frontmatter** — `--- key: value ---` config at the top of `.md` files

## Install

```bash
go install github.com/manic/goweb/cmd/goweb@latest
```

Or build from source:

```bash
git clone https://github.com/manic/goweb
cd goweb
go build -o goweb ./cmd/goweb
```

## Syntax

### Chunk definition

```
<<name>>= file: output.go pipe: gofmt
code body
spans multiple lines
>>
```

Attributes:
| Attribute | Description |
|-----------|-------------|
| `file: path` | Write resolved chunk to this file path |
| `pipe: cmd` | Pipe resolved content through an external command |

Chunks can appear inside fenced code blocks or directly in markdown.

### Chunk reference

```
<<name>>
```

References are expanded recursively during tangling.

### Conditional blocks

```
<<if debug>>
content for debug mode
<<elif verbose>>
content for verbose mode
<<else>>
content for other cases
<<end>>
```

Conditions are evaluated using `--var` flags. Supported expressions:
- `name` — true if the variable is set to a truthy value (not empty, `"false"`, `"0"`, or `"no"`)
- `name==value` — true if equality holds
- `name!=value` — true if inequality holds

### Cross-file import

```
<<import "other.md">>
```

Loads all chunks from `other.md`. Paths are resolved relative to the importing file.

## Usage

### Tangle (extract code)

```bash
# Tangle all chunks with file: attributes
goweb tangle --var debug=true program.md

# Tangle a specific chunk to stdout
goweb tangle program.md main > main.go

# Dry run (show what would be written)
goweb tangle --dry-run program.md

# Set pipe working directory
goweb tangle --pipe-dir ./scripts program.md
```

### Weave (generate documentation)

```bash
# Output clean markdown to stdout (chunk bodies wrapped in fenced code blocks)
goweb weave program.md

# Write to file
goweb weave program.md -o README.md

# With variables
goweb weave --var debug=false program.md
```

### Render (HTML documentation)

```bash
# Generate HTML with the built-in light documentation theme
goweb render program.md > docs.html

# With custom template
goweb render program.md --template custom.tmpl > docs.html

# With variables (expands {{var}} placeholders, controls <<if>> blocks)
goweb render program.md --var APP_NAME="My App" --var debug=true > docs.html

# Site mode — render a directory of .md files
goweb render --site docs/ --output-dir site/
```

The default theme is a **light** documentation layout in the spirit of [lit-lang/lit](https://github.com/lit-lang/lit) / GitBook docs:
- White page, light sidebar, system/Inter typography, blue accents
- **highlight.js** GitHub (light) syntax highlighting only
- Sidebar TOC, breadcrumbs, chunk index
- SEO (Open Graph, Twitter, canonical) and AEO (JSON-LD, semantic article)
- Code language auto-detected from `file:` extension (`.tsx` → TypeScript, `.css` → CSS, etc.)
- Responsive layout with mobile sidebar toggle

## Architecture

```
source.md
    │
    ▼
┌─────────────────────┐
│   Preprocessor      │  resolves <<import>>, <<if>>/<<elif>>/<<else>>/<<end>>
│   (pkg/preproc)     │  using --var flags, produces filtered lines
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│   Parser            │  identifies fenced code blocks, extracts <<name>>=..>>
│   (pkg/parser)      │  chunk definitions, collects <<ref>> references
└─────────┬───────────┘
          │
          ▼
┌─────────────────────┐
│   Dependency Graph  │  builds dependency graph, topological sort, cycle detect
│   (pkg/graph)       │
└─────────┬───────────┘
          │
          ├──► Tangle ──► output files
          │    (pkg/tangle)   (resolves refs, applies pipes)
          │
          ├──► Weave  ──► clean markdown
          │    (pkg/weave)    (strips goweb syntax, wraps code in fences)
          │
          └──► Render ──► HTML documentation
               (cmd/goweb)     (light theme, syntax highlighting)
```

## Example

There is a **single** example: a Preact + Bun + Tailwind TODO app (`examples/preact-todo/`).

A frontend TODO app built with **Preact**, **Bun**, and **Tailwind CSS** that demonstrates:
- `<<chunk>>=` definitions with `file:` → outputs to `src/*.tsx`, `package.json`, `index.html`, etc.
- `<<import "file.md">>` — cross-file references across 4 source files
- `<<if debug>>`/`<<end>>` — conditional console.log for debug mode
- `tags:` — categorization (e.g., `tags: component`, `tags: config`, `tags: debug`)
- `--var` — configurable `APP_NAME`, `AUTHOR`, and `debug` variables
- `{{var}}` substitution — variables in `package.json`, `index.html`, LICENSE
- `<<override>>` — replace components easily
- Multiple sub-components (App, TodoList, TodoItem, AddTodo)
- Tailwind CSS utility classes for responsive design

```bash
cd examples/preact-todo
goweb tangle --var APP_NAME="My Todo" --var AUTHOR="You" --var debug=true main.md
bun install
bun run dev
```

Render the documentation to HTML with the light theme:
```bash
goweb render main.md --var APP_NAME="Todo App" --var AUTHOR="You" --var debug=true -o index.html
```

The rendered HTML is published on GitHub Pages (`gh-pages`) with SEO and AEO metadata.

## Project Structure

```
cmd/goweb/            CLI (cobra): tangle, weave, render, index, graph, reverse, sync, lsp, init
pkg/parser/           Chunk, Document, markdown/noweb parser
pkg/preproc/          Imports, conditionals, frontmatter
pkg/graph/            Dependency graph, topological sort
pkg/tangle/           Reference expansion, pipes, exec, file output
pkg/weave/            Strip goweb syntax → clean markdown
examples/preact-todo/ Single example (Preact + Bun + Tailwind)
SPEC.md               Language and architecture spec (agents + humans)
```

## License

MIT
