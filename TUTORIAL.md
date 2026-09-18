# goweb Tutorial

A step-by-step guide to literate programming with goweb.

## Installation

```bash
go install github.com/xyzzyapps/goweb/cmd/goweb@latest
```

Or build from source:

```bash
git clone <repo> && cd goweb
go build -o goweb ./cmd/goweb
```

## 1. Your First Literate Program

Create a new project:

```bash
goweb init myproject
cd myproject
```

This creates `program.md` with a simple Go program. Look at the file:

````markdown
# My Literate Program

This is a goweb literate programming project.

<<package>>= file: main.go
package main

import "fmt"

func main() {
    fmt.Println("Hello from goweb!")
}
>>
````

The `<<package>>= file: main.go` line defines a chunk named "package" that writes to `main.go`. The body follows until `>>` on its own line.

Tangle it:

```bash
goweb tangle program.md
```

This generates `main.go` with the resolved code.

## 2. Chunks and References

Split your code into named chunks. A chunk can reference another using `<<name>>`:

````markdown
<<greeting>>=
"Hello, World!"
>>

<<main>>= file: main.go
package main

import "fmt"

func main() {
    fmt.Println(<<greeting>>)
}
>>
````

Tangling this:
- `<<greeting>>` is resolved to `"Hello, World!"`
- `<<main>>` assembles the final `main.go`

## 3. Out-of-Order Definitions

Chunks can appear in any order. goweb collects all chunks first, then resolves references:

````markdown
<<a>>=
<<b>>
<<c>>
>>

<<c>>=
world
>>

<<b>>=
hello
>>
````

Order doesn't matter — `<<a>>` resolves to `hello\nworld`.

## 4. Cross-File Imports

Split large projects across multiple `.md` files:

**types.md:**
````markdown
<<types>>=
type User struct {
    Name string
    Age  int
}
>>
````

**main.md:**
````markdown
<<import "types.md">>

<<main>>= file: main.go
package main

<<types>>

func main() {}
>>
````

Tangling `main.md` pulls in chunks from `types.md`.

## 5. Variables and Conditionals

Use `--var` to control what gets tangled:

**program.md:**
````markdown
<<imports>>=
import "fmt"
<<if debug>>
import "log"
<<end>>
>>

<<main>>= file: main.go
package main

<<imports>>

func main() {
    fmt.Println("Hello")
    <<if debug>>
    log.Println("debug mode")
    <<end>>
}
>>
````

```bash
# With debug:
goweb tangle --var debug=true program.md

# Without debug:
goweb tangle program.md
```

## 6. Inline Variable Substitution

Use `{{var}}` to substitute values inside code:

````markdown
<<config>>= file: config.json
{
    "host": "{{HOST}}",
    "port": {{PORT}}
}
>>
````

```bash
goweb tangle --var HOST=localhost --var PORT=5432 program.md
```

## 7. Code Execution

Run chunk bodies through interpreters with `exec:`:

````markdown
<<example>>= exec: python3
print("Hello from Python!")
>>
````

During tangling, the body is piped through `python3` and stdout replaces the chunk content.

## 8. Session-Based Execution

Multiple chunks can share the same interpreter session:

````markdown
<<setup>>= exec: python3 session: calc
x = 42
>>

<<compute>>= exec: python3 session: calc
print(f"x = {x}")
>>
````

The second chunk reuses the Python process from the first, preserving state.

## 9. Pipes

Transform code with `pipe:` attributes:

````markdown
<<code>>= file: out.go pipe: gofmt
package main
func main(){}
>>
````

The body is piped through `gofmt` before being written.

## 10. Chunk Overriding

Override previous definitions with the `override` keyword:

````markdown
<<default>>= file: config.yaml
host: localhost
port: 8080
>>

<<override "custom.md">>

<<default>>= override
host: {{HOST}}
port: {{PORT}}
>>
````

## 11. Source Line Directives

Track where code originates:

```bash
goweb tangle --line-directives program.md
```

The output includes `//line` or `#line` comments pointing back to the `.md` source:

```go
//line "program.md":3
package main
```

## 12. Output Directory

Keep tangled output separate from source:

```bash
goweb tangle --output-dir ./build program.md
goweb weave --output-dir ./docs program.md
```

## 13. Watch Mode

Auto-tangle when the source file changes:

```bash
goweb tangle --watch program.md
```

The tool watches for saves and re-tangles automatically.

## 14. Weave (Documentation)

Generate clean markdown from your literate program:

```bash
goweb weave program.md > README.md
```

This strips control syntax. In prose, `<<name>>` becomes a link to that chunk; reserved names (`override`, `import`, …) stay literal.

## 15. Render, index, and graph

```bash
goweb render program.md -o docs.html
goweb index program.md
goweb graph program.md | dot -Tsvg -o deps.svg
goweb tangle --match component program.md
```

Unset `AUTHOR`, `EMAIL`, and `REPO` come from `git config`. See [SPEC.md](SPEC.md) and the example book in `examples/preact-todo/docs.html`.

## 16. AI holes

Mark a stub `ai:open`. Fill rewrites the **Markdown**. Then seal so later fills skip it.

```
<<add>>= file: add.js ai:open
function add() {}
>>
```

```bash
goweb fill --dry-run program.md
goweb fill --mock program.md
goweb seal program.md
```

`--mock` reads `mock/<chunk-name>.txt` next to the source file. Unmarked chunks are never filled.

## Summary

| Command | Description |
|---------|-------------|
| `goweb init [dir]` | Scaffold a new project |
| `goweb tangle file.md` | Extract code chunks into source files |
| `goweb tangle file.md chunk` | Print a single chunk to stdout |
| `goweb tangle --watch file.md` | Auto-tangle on file changes |
| `goweb tangle --match tag` | Only chunks with that tag |
| `goweb weave file.md` | Generate clean markdown docs |
| `goweb render file.md` | HTML documentation |
| `goweb index file.md` | Cross-reference table |
| `goweb graph file.md` | Graphviz DOT dependency graph |
| `goweb sync file.md` | Apply tangled edits back (needs `--line-directives`) |
| `goweb reverse files-or-dirs…` | Wrap source files (or a folder) as chunks |
| `goweb fill file.md` | Fill `ai:open` chunks (`--mock` or a model; `--dry-run` lists) |
| `goweb seal file.md` | `ai:filled` → `ai:sealed` |
| `goweb tangle --var key=val` | Set variables for conditionals/`{{var}}` |
| `goweb tangle --line-directives` | Add source-location comments |
| `goweb tangle --output-dir ./out` | Write output to a subdirectory |
| `goweb tangle --pipe-dir ./scripts` | Working dir for pipe/exec commands |
