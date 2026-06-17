# goweb TODO App Example

This example demonstrates most goweb features in a working TODO application:

| Feature | Demonstrated |
|---|---|
| `<<chunk>>=` definitions | Every code block |
| `<<ref>>` references | `<<types>>`, `<<methods>>`, `<<imports>>`, etc. |
| `<<import "file.md">>` | Loading `types.md` and `methods.md` |
| `<<if>>`/`<<end>>` conditionals | Debug mode toggling |
| `--var debug=true\|false` | Controlling debug output |
| `file: path` attribute | `main.go` and `LICENSE` output files |
| `tags:` attribute | `tags: debug`, `tags: meta` on chunks |
| `{{var}}` substitution | `{{YEAR}}` in license |
| Cross-file references | `<<types>>` from `types.md`, `<<methods>>` from `methods.md` |

## Usage

```bash
# Production build (no debug logging)
goweb tangle --var debug=false main.md
go build -o todo main.go

# Debug build (with extra logging)
goweb tangle --var debug=true main.md --var YEAR=2026
go build -o todo-debug main.go

# Generate cross-reference index
goweb index main.md

# Generate dependency diagram
goweb graph main.md | dot -Tpng -o deps.png

# Render to HTML
goweb render main.md > todo.html

# Only tangle debug chunks
goweb tangle --match debug main.md
```
