# Preact TODO App

<<import "app.md">>
<<import "components.md">>
<<import "config.md">>

This literate program demonstrates a **Preact + Bun + Tailwind** TODO application
while showcasing nearly all of goweb's features. Every source file, config file,
and build script is defined as a named chunk inside this markdown document.

## Goweb features demonstrated

| Feature | Usage in this example |
|---|---|
| `<<chunk>>=` definitions | Every source file is a named chunk with a `file:` attribute |
| `<<import "file.md">>` | Cross-file references between all `.md` files |
| `<<ref>>` references | Handler stubs defined as named refs, resolved during tangle |
| `<<if>>`/`<<else>>`/`<<end>>` | Conditional debug logging via `--var debug=true\|false` |
| `tags:` attribute | Chunks tagged `component`, `config`, `debug`, `meta` |
| `file:` attribute | Each chunk maps directly to a source or config file |
| `--var` variables | `APP_NAME`, `AUTHOR`, `debug` injected at tangle/render time |
| `{{var}}` substitution | `{{APP_NAME}}`, `{{AUTHOR}}` placeholders in JSON, HTML, LICENSE |
| Language inference | `.tsx` → TypeScript, `.css` → CSS, `.json` → JSON in rendered output |
| `goweb tangle` | Extract all chunks into runnable source files |
| `goweb render` | Generate full HTML documentation with the light theme |
| `goweb weave` | Produce clean markdown with fenced code blocks |
| Chunk order independence | Handler functions defined after their usage in the render tree |
| Cross-file chunk references | Components imported from `components.md`, config from `config.md` |

## Usage

```bash
# Tangle all source files into the examples/preact-todo/ directory
goweb tangle --var APP_NAME="Todo App" --var AUTHOR="You" main.md

# With debug logging (adds console.log statements)
goweb tangle --var debug=true main.md

# Install dependencies and run
bun install
bun run dev

# Generate HTML documentation with the light theme
goweb render main.md --var debug=true --var APP_NAME="Todo App" --var AUTHOR="You" -o index.html
```

## How the chunks are organized

The example is split across four files:

- **`main.md`** (this file) — entry point with imports, overview, tangle/usage instructions, and the LICENSE chunk
- **`app.md`** — the main `App` component with state management, filter buttons, and conditional debug logging
- **`components.md`** — three presentational components (`AddTodo`, `TodoList`, `TodoItem`) with Tailwind styling
- **`config.md`** — build configuration files (`package.json`, `tsconfig.json`, `tailwind.config.js`, `index.html`)

Each file exports named chunks that are resolved during tangling. The `<<import "file.md">>>` directive merges the imported file's chunks into the namespace, so chunks defined in `components.md` can be referenced from `app.md`.

## Chunk overriding

The `TodoList` component chunk is tagged with `override-demo`. If you define a chunk with the same name
in another file and mark it with `<<override>>`, goweb will replace the original definition.
This is useful for swapping implementations without modifying the original source:

```bash
# Create an override chunk
echo '<<todo-list-component>>= file: src/components/todo-list.tsx
export function TodoList({ todos }) {
  return <div>Custom implementation</div>;
}
>>' > override.md

# Tangle with override
goweb tangle --var APP_NAME="Todo App" main.md override.md
```

## Project Structure

```
examples/preact-todo/
├── main.md             # Entry point (this file)
├── app.md              # App component definition
├── components.md       # UI sub-components
├── config.md           # Build config files
├── package.json        # Tangled output — Bun project
├── tsconfig.json       # TypeScript configuration
├── tailwind.config.js  # Tailwind CSS theme
├── index.html          # HTML entry point
└── src/
    ├── main.tsx        # Preact mount point
    ├── app.tsx         # App component
    ├── style.css       # Tailwind directives
    └── components/
        ├── add-todo.tsx
        ├── todo-list.tsx
        └── todo-item.tsx
```

## Conditional compilation with <<if>>

The `app.md` file uses `<<if debug>>` / `<<end>>` directives to conditionally include
`console.log` statements. When `--var debug=true` is passed, the debug chunks are included;
otherwise they are stripped. This allows a single source to produce both production
and development builds.

## Tags and chunk metadata

Chunks can be tagged with `tags:` for organization and filtering:

- **`component`** — UI components tangled to `src/components/`
- **`config`** — build configuration files
- **`debug`** — conditional debug logging chunks
- **`meta`** — metadata files like `LICENSE`

The `goweb graph` command can display relationships between tagged chunks.

## License

<<license>>= file: LICENSE tags: meta
MIT License

Copyright (c) 2026 {{APP_NAME}}

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
>>
