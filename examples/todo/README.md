# TODO App — goweb Literate Programming Example

This example demonstrates all goweb features:

| Feature | How |
|---|---|
| **Noweb-style chunks** | `<<name>>=` ... `>>` |
| **References** | `<<name>>` inside code expands the chunk |
| **Out-of-order definitions** | Chunks defined anywhere; goweb resolves all refs |
| **Cross-file references** | `<<import "file.md">>` loads chunks from other files |
| **Debug toggling** | `<<if debug>>` / `<<end>>` controlled by `--var debug=true|false` |
| **File output** | `file: main.go` on a chunk writes to that path |
| **Pipe transform** | `pipe: gofmt` formats the generated code |

## Usage

```bash
# Production build (no debug)
goweb tangle --var debug=false main.md
go build -o todo main.go

# Debug build (extra logging)
goweb tangle --var debug=true main.md
go build -o todo-debug main.go
```

## File structure

```
main.md      — entry point, imports other files, main loop
types.md     — data types
methods.md   — methods on Todos
README.md    — this file
```

`main.md` imports `types.md` and `methods.md` via `<<import>>`.
Chunks across all files are merged into a single namespace.
The `file: main.go` on `<<package>>` means tangled output goes to `main.go`.
