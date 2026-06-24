# Preact TODO App

<<import "app.md">>
<<import "components.md">>
<<import "config.md">>

This example demonstrates a TODO application built with **Preact**, **Bun**, and **Tailwind CSS**,
structured as a literate program using goweb's chunk syntax.

Each component is defined as a named chunk with a `file:` attribute for tangling.

| Feature | How Demonstrated |
|---|---|
| `<<chunk>>=` definitions + `file:` | Each component tangled to `src/*.tsx` |
| `<<import "file.md">>` | Cross-file refs between all `.md` files |
| `<<if debug>>`/`<<end>>` | Conditional console.log for state changes |
| `tags:` attribute | Chunks tagged `component`, `config`, `debug` |
| `--var app-name` | Project name in `package.json` |
| `--var debug=true\|false` | Toggle debug logging |
| `{{var}}` substitution | `{{APP_NAME}}`, `{{AUTHOR}}` in config |
| `<<override>>` | Replace components easily |

## Usage

```bash
# Tangle all source files
goweb tangle --var APP_NAME="Todo App" --var AUTHOR="You" main.md

# With debug logging
goweb tangle --var debug=true main.md

# Run with bun
bun install
bun run dev
```

## Project Structure

```
examples/preact-todo/
├── main.md             # Entry point (this file)
├── app.md              # App component definition
├── components.md       # UI sub-components
├── config.md           # Build config files
├── package.json        # Tangled output — Bun project
├── tsconfig.json       # TypeScript config
├── tailwind.config.js  # Tailwind CSS config
├── index.html          # HTML entry point
└── src/
    ├── main.tsx        # Preact mount point
    ├── app.tsx         # App component
    ├── style.css       # Tailwind imports
    └── components/
        ├── add-todo.tsx
        ├── todo-list.tsx
        └── todo-item.tsx
```

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
