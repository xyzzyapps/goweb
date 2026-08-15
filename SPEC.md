# goweb specification

This document is the source of truth for **humans and agents** working on goweb. It describes the language, pipeline, packages, CLI, examples, and invariants. Prefer this file over comments or chat history when behavior is unclear.

**Authorship.** The original implementation was written with **Grok**, **DeepSeek**, **Flask**, and **Gemini**. Subsequent work in this tree (this spec, theme, docs, history cleanup, and publishing) was done by **Grok 4.6** (xAI).

---

## 1. Purpose

goweb is a noweb-style literate programming tool on GitHub Flavored Markdown. A program is one or more `.md` files that mix prose and named code chunks. Operators:

| Command | Role |
|---------|------|
| `tangle` | Resolve chunks and write source files |
| `weave` | Strip goweb syntax → clean markdown |
| `render` | Weave + Goldmark → HTML (default light theme) |
| `sync` | Apply edits from tangled files back into `.md` (needs `#line` / `//line`) |
| `index` | Cross-reference table of chunks |
| `graph` | Graphviz DOT dependency graph |
| `reverse` | Wrap existing source files as chunks |
| `init` | Scaffold `program.md` |
| `lsp` | JSON-RPC language server over stdin/stdout |

Module path: `github.com/manic/goweb`. Go version: see `go.mod`.

---

## 2. Source language

### 2.1 YAML frontmatter

Optional block at the top of a file:

```
---
key: value
---
```

`pkg/preproc` strips it and may merge keys into the variable map used by conditionals and `{{var}}`.

### 2.2 Chunk definition

```
<<name>>= file: out.go pipe: gofmt exec: python session: repl tags: a,b override
body
>>
```

- Header is `<<name>>=` optionally followed by attributes.
- Terminator is `>>` alone on a line (or escaped `\>>` as literal `>>` in the body).
- Definitions may sit **inside** fenced ```` ``` ```` / `~~~` blocks **or** as noweb-style blocks in prose.
- Same name: bodies are concatenated unless `override` / `<<override "file.md">>` replaces the previous definition.

**Attributes**

| Attribute | Meaning |
|-----------|---------|
| `file: path` | Tangled output path (relative to `--output-dir` if set) |
| `pipe: cmd` | Pipe resolved text through an external command |
| `exec: cmd` | Run body through interpreter; stdout replaces body |
| `session: name` | Share an exec process across chunks (see SessionManager) |
| `tags: a,b` | Labels; `--match` filters tangle by tag |
| `override` | Replace any earlier chunk of the same name |

Language for weave/render fences: fenced info string, else inferred from `file:` extension (`.tsx` → TypeScript, `.css` → CSS, …).

### 2.3 Chunk reference

`<<name>>` inside a body is expanded recursively at tangle time. Definitions (`<<name>>=`) are not treated as references. Cycles are errors (`pkg/graph`).

### 2.4 Imports

```
<<import "other.md">>
<<override "other.md">>
```

Paths are relative to the importing file. Imports are inlined by the preprocessor. Override imports inject a `<<__goweb_override__>>` marker so following chunks replace existing names. Duplicate import of the same absolute path is a no-op (cycle/idempotence).

### 2.5 Conditionals

```
<<if expr>>
...
<<elif expr>>
...
<<else>>
...
<<end>>
```

Evaluated in `pkg/preproc` **before** parse.

Expressions:

- `name` — truthy unless unset, empty, `"false"`, `"0"`, or `"no"`
- `name==value` / `name!=value`

Variables come from `--var key=value` (bare `--var debug` ⇒ `"true"`). Nested conditionals are supported; skipped parents skip children.

### 2.6 Inline variables

`{{NAME}}` is substituted from the same var map during tangle (graph resolver) and render (markdown pass). Unknown placeholders stay as-is.

---

## 3. Pipeline

```
.md  →  preproc (import, if/elif/else/end, frontmatter)
     →  parser (fences + <<name>>= … >>)
     →  graph (refs, topo sort, cycles)
     →  tangle | weave | render | index | graph | sync
```

**Invariants**

1. Preprocessor is the only stage that follows imports and evaluates `<<if>>`.
2. Parser never executes code; it only builds `[]*Chunk`.
3. Tangle writes only chunks with `file:` (unless `TangleChunk` is asked for a name).
4. Weave never writes binary; it produces markdown.
5. Render weaves first, then Goldmark (GFM + auto heading IDs + unsafe HTML).
6. `sync` is only reliable if tangle was run with `--line-directives`.

---

## 4. Packages

### `pkg/parser`

- `Chunk`, `Document`, `Merge`, tags, `ChunkTable`.
- `ParseLines(lines, sourcePath)` — fence state machine + noweb chunk state.
- Duplicate names across files: `Merge` unless override.

### `pkg/preproc`

- `Process(root, vars) → Result{Lines, Sources}`.
- Conditional stack: `StateNormal | Active | Skipping | Done`.
- Import/override path resolution.

### `pkg/graph`

- Builds name → deps from `<<ref>>`.
- Kahn topological sort; cycle error.
- Used by tangle resolution.

### `pkg/tangle`

- `Tangle` config: `OutputDir`, `PipeDir`, `DryRun`, `LineDirectives`, `MatchTags`.
- Expand refs, `{{var}}`, exec/pipe, write files.
- `SessionManager` for `session:` (persistent stdin process).
- Line directives: `//line "file":N` (Go) or `#line N "file"` (C-like).

### `pkg/weave`

- Preprocess, then `stripControlSyntax`: drop headers/`>>`/directives; wrap implicit chunk bodies in fences.

### `cmd/goweb`

Cobra root. Subcommands live in:

| File | Command |
|------|---------|
| `main.go` | `tangle`, `weave`, `init`, flags, watch |
| `render_cmd.go` | `render`, default HTML theme, site mode |
| `index_cmd.go` | `index` |
| `graph_cmd.go` | `graph` (`--cluster` by output file) |
| `reverse_cmd.go` | `reverse` |
| `sync_cmd.go` | `sync` |
| `lsp_cmd.go` | `lsp` |

Dependencies: `spf13/cobra`, `fsnotify`, `yuin/goldmark`.

---

## 5. CLI contract

```
goweb tangle [flags] <source.md> [chunk]
  --var/-v, --pipe-dir/-P, --dry-run/-n, --line-directives
  --watch/-w, --output-dir/-O, --match/-m

goweb weave [flags] <source.md>
  --var, --output/-o, --output-dir/-O

goweb render <source.md> [...]
  --template/-t, --output/-o, --site/-s, --output-dir/-O, --var
  Site mode requires --output-dir.

goweb index <source.md>   --var
goweb graph <source.md>   --var --cluster
goweb reverse <files...>  --output
goweb sync <source.md>    --var --dry-run
goweb init [directory]
goweb lsp
```

Watch mode re-tangles on write/create (100ms debounce).

---

## 6. Default HTML theme

When `--template` is omitted, `renderDefaultTheme` emits a **light** documentation page (GitBook / [lit-lang/lit](https://github.com/lit-lang/lit) docs style):

- White page, light sidebar, dark text, blue accents
- System / Inter-like sans, no dark-mode media query
- highlight.js `github` style only
- Sidebar TOC from headings; chunk index `<details>`
- SEO: `description`, Open Graph, Twitter card, canonical
- AEO: JSON-LD `SoftwareApplication` + `TechArticle`, semantic `article`, `llms.txt` expected on published sites

Do **not** reintroduce the Read the Docs dark navy sidebar or `prefers-color-scheme: dark` as default.

Template context (`PageData`): `Title`, `Content`, `Source`, `Chunks`, `Headings`.

---

## 7. Example

A **single** example lives in `examples/preact-todo/`:

- Literate sources: `main.md`, `app.md`, `components.md`, `config.md`
- Tangle → Preact + Bun + Tailwind app (`src/*`, `package.json`, `index.html`, …)
- Tests: `preact_todo_test.go` (tangle smoke)
- Demo vars: `APP_NAME`, `AUTHOR`, `debug`

```bash
cd examples/preact-todo
goweb tangle --var APP_NAME="Todo App" --var AUTHOR="You" --var debug=true main.md
bun install && bun run dev
goweb render main.md --var APP_NAME="Todo App" --var AUTHOR="You" --var debug=true -o index.html
```

Do not add a second example tree unless the product owner asks. The old Go CLI `examples/todo/` was removed.

---

## 8. Tests and fixtures

- Unit tests next to packages: `pkg/*/*_test.go`, `cmd/goweb/render_cmd_test.go`
- `testdata/` for parser/tangle fixtures
- Example test must keep working after markdown edits
- `go test ./...` is the gate before publish

---

## 9. Publishing

- **Source:** GitHub repo `goweb` on the authenticated `gh` account, default branch `master`
- **Docs/demo:** `gh-pages` branch: rendered example `index.html` plus `llms.txt`, `robots.txt`, `sitemap.xml`
- Do not commit `.todo/` or `TODO.md` (scrubbed from history)

---

## 10. Agent working rules

1. Change language behavior only with tests.
2. Keep chunk syntax compatible with existing examples.
3. Theme edits stay in `renderDefaultTheme` unless a custom `--template` path is added.
4. Never restore RTD dark chrome or a second example without being asked.
5. After UI/theme changes, re-render the example HTML.
6. `go test ./...` and tangle the example before claiming done.
7. History must stay free of `.todo/` and `TODO.md`.
