# goweb specification

This document is the source of truth for **humans and agents** working on goweb. It describes the language, pipeline, packages, CLI, examples, and invariants. Prefer this file over comments or chat history when behavior is unclear.

**Authorship.** This codebase was written with **DeepSeek** and **Grok 4.6** (xAI).

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

Module path: `github.com/xyzzyapps/goweb`. Go version: see `go.mod`.

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

When unset, the CLI fills `AUTHOR` from `git config user.name`, `EMAIL` from `user.email`, and `REPO` from `remote.origin.url` (https, `.git` stripped). Explicit `--var` values are not overwritten.

### 2.6 Inline variables

`{{NAME}}` is substituted from the same var map during tangle (graph resolver) and render (markdown pass). Unknown placeholders stay as-is.

### 2.7 Reserved `<<…>>` names

These are **not** chunk references. Weave leaves them as literal `<<name>>` (and does not emit `#chunk-…` links):

`if`, `elif`, `else`, `end`, `import`, `override`, `__goweb_override__` (and `import …` / `override …` with a path).

`<<name>>` already inside markdown backticks is left alone so `` `<<override>>` `` stays a code span.

### 2.8 Weave / render cross-references

In **prose**, a non-reserved `<<name>>` becomes an HTML link to `#chunk-…` (`parser.ChunkAnchor`). Each noweb-style definition gets that anchor and a heading. If a chunk body references other chunks, weave appends a **Uses …** line after the fenced block. The render theme’s chunk table links to the same ids.

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
4. Weave never writes binary; it produces markdown. Prose `<<name>>` becomes a link to `#chunk-…`; each definition gets that anchor. Chunk bodies that reference other chunks get a “Uses …” line.
5. Render weaves first, then Goldmark (GFM + auto heading IDs + unsafe HTML).
6. `sync` is only reliable if tangle was run with `--line-directives`.

---

## 4. Packages

### `pkg/parser`

- `Chunk`, `Document`, `Merge`, tags, `ChunkTable`, `ChunkAnchor`.
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

- Preprocess, then `stripControlSyntax`: drop headers/`>>`/directives; wrap implicit chunk bodies in fences; linkify prose refs; **Uses** lines; skip reserved names.

### `cmd/goweb`

Cobra root. Subcommands live in:

| File | Command |
|------|---------|
| `main.go` | `tangle`, `weave`, `init`, flags, watch, `parseVars` |
| `gitvars.go` | `AUTHOR`/`EMAIL`/`REPO` from `git config` when unset |
| `theme.go` | `renderDefaultTheme` (lit-lang.org palette, no link underlines) |
| `render_cmd.go` | `render`, site mode, `PageData` |
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

When `--template` is omitted, `renderDefaultTheme` matches [lit-lang.org](https://lit-lang.org/) **without** link or heading underlines:

- System UI font stack, `#ffcd42` primary, `#291f04` footer, `#333` body text
- 2px black borders and 4px offset shadows on tables and code
- Centered header (logo + Source / Documentation) and `main` (max-width 800px)
- Links: bold, no underline; hover fades opacity
- SEO: `description`, Open Graph, Twitter card, canonical (`Pages` from `REPO`)
- AEO: JSON-LD `SoftwareApplication` + `TechArticle`; `llms.txt` on published sites
- Footer / meta author from git-backed `AUTHOR` / `EMAIL` / `REPO` when set

Do **not** reintroduce a Read the Docs navy sidebar, auto dark mode, or underlined links. The example todo app uses the same palette. Header: **View source** → `docs.html`, **GitHub** → `{{REPO}}`.

Template context (`PageData`): `Title`, `Content`, `Source`, `Chunks` (incl. `ID`), `Headings`, `Author`, `Email`, `Repo`, `Pages`.

---

## 7. Example

A **single** example lives in `examples/preact-todo/`:

- Literate sources: `main.md`, `app.md`, `components.md`, `config.md` (also linked from rendered docs)
- Tangle → Preact + htm ES modules (`src/*.js`, `package.json`, `index.html`) — **no bundler**, no Tailwind/Bun required to run
- `docs.html` — `goweb render` of `main.md` (literate programming primer, tags/graphs/LLM context)
- Tests: `preact_todo_test.go` (tangle smoke)
- Vars: `APP_NAME` via `--var`; `AUTHOR`/`EMAIL`/`REPO` from git config unless overridden; `debug` for logs
- Tags: `component`, `config`, `debug`, `meta`, `override-demo`

```bash
cd examples/preact-todo
goweb tangle --var APP_NAME="Todo App" --var debug=true main.md
python -m http.server 8080
goweb render main.md --var APP_NAME="Todo App" --var debug=true -o docs.html
```

Do not add a second example tree unless the product owner asks. The old Go CLI `examples/todo/` was removed.

---

## 8. Tests and fixtures

- Unit tests next to packages: `pkg/*/*_test.go`, `cmd/goweb/*_test.go`, `examples/preact-todo/preact_todo_test.go`
- Example test must keep working after markdown edits
- `go test ./...` is the gate before publish

---

## 9. Publishing

- **Source:** `github.com/xyzzyapps/goweb`, default branch `master`
- **License:** Creative Commons Attribution-ShareAlike 4.0 (`LICENSE`)
- **Docs/demo:** `gh-pages` — example `index.html` (app), `docs.html`, `llms.txt`, `robots.txt`, `sitemap.xml`
- Do not commit `.todo/`, `TODO.md`, or `draft/` (gitignore; `.todo` was scrubbed from history)
- Agent skill: `.agents/skills/goweb/SKILL.md`

---

## 10. Agent working rules

1. Change language behavior only with tests.
2. Keep chunk syntax compatible with existing examples.
3. Theme edits stay in `renderDefaultTheme` unless a custom `--template` path is added.
4. Never restore RTD dark chrome or a second example without being asked.
5. After UI/theme changes, re-render the example HTML.
6. `go test ./...` and tangle the example before claiming done.
7. History must stay free of `.todo/` and `TODO.md`. Do not commit `draft/`.
8. `<<override>>` is a directive, never a chunk link.
