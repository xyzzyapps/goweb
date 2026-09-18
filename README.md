# goweb

**goweb** is a literate programming tool that brings noweb-style chunk syntax to GitHub Flavored Markdown. Write documentation and code in a single `.md` file, then **tangle** to extract source files, **weave** to generate clean markdown, or **sync** to apply edits back from the generated code.

This codebase was written with **DeepSeek** and **Grok 4.6** (xAI). See [SPEC.md](SPEC.md) for the full language and architecture contract (for humans and agents).

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
- **Weave** — clean markdown; prose `<<name>>` becomes in-document chunk links
- **Render** — HTML (lit-lang.org palette, no underlines) via `goweb render`
- **Cross-reference index** — `goweb index` prints a table of all chunks and their references
- **Dependency diagrams** — `goweb graph` outputs Graphviz DOT (`--cluster` by output file)
- **Tags** — `tags:` shelves; `goweb tangle --match` extracts one shelf (good LLM context with `index`/`graph`)
- **Source → literate** — `goweb reverse` converts files or a folder into `.md` chunks
- **Project scaffolding** — `goweb init` creates a new literate program skeleton
- **YAML frontmatter** — `--- key: value ---` config at the top of `.md` files
- **AI holes** — `ai:open` → `goweb fill` → `ai:filled` → `goweb seal` → `ai:sealed`

## Install

```bash
go install github.com/xyzzyapps/goweb/cmd/goweb@latest
```

Or build from source:

```bash
git clone https://github.com/xyzzyapps/goweb
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
| `exec: cmd` | Run the body; stdout replaces the chunk |
| `session: name` | Share an exec process across chunks |
| `tags: a,b` | Labels; `goweb tangle --match` filters by tag |
| `override` | Replace any earlier chunk of the same name |
| `ai:open` / `filled` / `sealed` | Hole for `goweb fill` / `goweb seal` |

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

### Graph and index (chunk map)

Tags (`tags: component`) are shelves. `goweb tangle --match component` writes only that shelf. The graph is who points at whom — useful for humans and as compact context for an LLM (map first, then one chunk).

```bash
# Text table: each chunk, where it is defined, who references it
goweb index program.md

# Graphviz DOT — boxes and arrows (A → B means A contains <<B>>)
goweb graph program.md | dot -Tsvg -o deps.svg
goweb graph --cluster program.md | dot -Tpng -o deps.png
```

`--cluster` groups chunks by the file they tangle into.

### Fill and seal (AI holes)

Only chunks with `ai:open` are fill candidates. Tangle never calls a model. Fill never sets `ai:sealed`.

```
<<add>>= file: add.js ai:open
function add() {}
>>
```

```bash
goweb fill --dry-run program.md          # list holes; no HTTP, no writes
goweb fill --mock program.md             # mock/<chunk>.txt
goweb fill --model grok-4.5 program.md   # OpenAI-compatible Chat Completions
goweb seal program.md                    # ai:filled → ai:sealed
```

Keys from `OPENAI_API_KEY` / `XAI_API_KEY` / `GOWEB_API_KEY` (and matching `*_BASE_URL` / `*_MODEL`).

### Reverse (files or a folder)

```bash
goweb reverse main.go lib.go > program.md
goweb reverse src/ -o program.md
```

The default theme follows [lit-lang.org](https://lit-lang.org/):
- System UI fonts, amber `#ffcd42` primary, dark footer, no link underlines
- Hard offset shadows on tables, asides, and code blocks
- Centered header + `main` (max 800px)
- SEO (Open Graph, Twitter, canonical) and AEO (JSON-LD, semantic article)
- Unset `AUTHOR` / `EMAIL` / `REPO` come from `git config`

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
          │    (pkg/tangle)   (resolves refs, applies pipes; never calls a model)
          │
          ├──► Weave  ──► clean markdown
          │    (pkg/weave)
          │
          ├──► Render ──► HTML documentation
          │    (cmd/goweb)
          │
          └──► Fill / seal ──► rewrite ai:* chunks in the .md
               (pkg/fill)
```

## Example

There is a **single** example: a Preact + htm TODO app (`examples/preact-todo/`) that runs in the browser with no bundler.

It demonstrates:
- `<<chunk>>=` definitions with `file:` → outputs to `src/*.js`, `package.json`, `index.html`, etc.
- `<<import "file.md">>` — cross-file references across 4 source files
- `<<if debug>>`/`<<end>>` — conditional console.log for debug mode
- `tags:` — shelves such as `component`, `config`, `debug`; `goweb tangle --match` and `goweb graph` / `goweb index` for maps and LLM context
- `--var` — configurable `APP_NAME`, `AUTHOR`, and `debug` variables
- `{{var}}` substitution — variables in `package.json`, `index.html`, LICENSE
- `<<override>>` — replace components easily
- Multiple sub-components (App, TodoList, TodoItem, AddTodo)
- Static ES modules + import map (serve the folder; do not open `file://`)

```bash
cd examples/preact-todo
goweb tangle --var APP_NAME="My Todo" --var debug=true main.md
python -m http.server 8080
```

Render the documentation to HTML with the light theme:
```bash
goweb render main.md --var APP_NAME="Todo App" --var debug=true -o docs.html
```

The rendered HTML is published on GitHub Pages (`gh-pages`) with SEO and AEO metadata.

## Project Structure

```
cmd/goweb/            CLI (cobra): tangle, weave, render, index, graph, reverse, fill, seal, sync, lsp, init
pkg/parser/           Chunk, Document, markdown/noweb parser
pkg/preproc/          Imports, conditionals, frontmatter
pkg/graph/            Dependency graph, topological sort
pkg/tangle/           Reference expansion, pipes, exec, file output
pkg/fill/             AI holes: fill (mock or HTTP) and seal
pkg/weave/            Strip goweb syntax → clean markdown
examples/preact-todo/ Single example (Preact + htm, no bundler)
SPEC.md               Language and architecture spec (agents + humans)
.agents/skills/goweb/ Agent skill for using this tool
```

## License

[Creative Commons Attribution-ShareAlike 4.0 International](https://creativecommons.org/licenses/by-sa/4.0/) (CC BY-SA 4.0). See [LICENSE](LICENSE).
