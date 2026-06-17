# goweb

**goweb** is a literate programming tool that brings noweb-style chunk syntax to GitHub Flavored Markdown. Write documentation and code in a single `.md` file, then **tangle** to extract source files or **weave** to generate clean markdown.

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
- **`--output-dir`** — base directory for all tangled/weave output files
- **Source line directives** — `--line-directives` inserts `//line`/`#line` comments pointing to the `.md` source
- **Watch mode** — `--watch` re-tangles automatically when the source file changes
- **Language auto-detect** — infers fenced-block language from `file:` extension
- **Escape `\>>`** — literal `>>` inside chunk bodies
- **Topological sorting** — chunks are emitted in dependency order
- **Cycle detection** — circular references are caught and reported
- **Weave** — strips goweb syntax to produce clean markdown
- **Project scaffolding** — `goweb init` creates a new literate program skeleton

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
# Output clean markdown to stdout
goweb weave program.md

# Write to file
goweb weave program.md -o README.md

# With variables
goweb weave --var debug=false program.md
```

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
          └──► Weave  ──► clean markdown
               (pkg/weave)    (strips goweb syntax)
```

## Example

See `examples/todo/` for a complete TODO app that demonstrates:
- Chunk definitions with `file:` and `pipe:` attributes
- Cross-file references via `<<import>>`
- Debug mode toggling with `<<if debug>>`/`<<end>>`
- Out-of-order chunk definitions
- Multiple output files from a single source

```bash
cd examples/todo
goweb tangle --var debug=true main.md
go build -o todo main.go
```

## Project Structure

```
cmd/goweb/main.go     CLI entry point (cobra)
pkg/parser/chunk.go   Chunk, Document data structures
pkg/parser/parser.go  Markdown parser, chunk extraction
pkg/preproc/preproc.go Preprocessor (imports, conditionals)
pkg/graph/graph.go    Dependency graph, topological sort, resolver
pkg/tangle/tangle.go  Tangle engine (reference expansion, pipes, file output)
pkg/weave/weave.go    Weave (strip goweb syntax, clean markdown)
examples/todo/        Complete TODO app example
```

## License

MIT
