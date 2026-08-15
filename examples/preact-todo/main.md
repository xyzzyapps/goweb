# A small todo, written as a book

<<import "app.md">>
<<import "components.md">>
<<import "config.md">>

You are reading a program the way you would read an essay.

The pages below are ordinary prose. Nested in them are named pieces of source —
*chunks* — that goweb can lift out into real files. The same text is both the
explanation and the thing being explained. That old idea is called
**literate programming**. Donald Knuth coined the phrase so a program could be
told as a story to a human first, and only later assembled for the machine.

If you have used a [Jupyter notebook](https://jupyter.org/), you already know a
cousin of this. A notebook interleaves markdown cells with code cells you run
in order. You scroll, you read, you execute. Literate programming is the same
impulse turned inside out:

- A notebook is a *session*. Cells run top to bottom; the document is a log of
  thinking in time.
- A literate program is a *book*. Chunks can appear in the order that teaches
  best. The tool later *tangles* them into the order the computer needs.

You do not press “Run cell.” You press `goweb tangle`, and a folder of
JavaScript appears. You press `goweb render`, and you get this HTML — the book
form. The handwritten pages are still here beside it:
[main.md](main.md), [app.md](app.md), [components.md](components.md),
[config.md](config.md). One source, two readings.

**On this page:** [What you are looking at](#what-you-are-looking-at) · [What Preact is](#what-preact-is) · [How the app works](#how-the-app-works) · [Usage](#usage) · [Tags, graphs, and talking to a model](#tags-graphs-and-talking-to-a-model) · [License](#license)

**Elsewhere:** [the running app](index.html) · [this page as markdown](main.md) · [app.md](app.md) · [components.md](components.md) · [config.md](config.md) · [goweb on GitHub]({{REPO}}) · [Preact](https://preactjs.com/) · [htm](https://github.com/developit/htm) · [lit-lang.org](https://lit-lang.org/) · [CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/)

## What you are looking at

This particular book builds a tiny **todo list** you can open in a browser:
type a task, tick it done, hide the finished ones, throw one away. There is a
**View source** button at the top of the app that comes back here, and a
**GitHub** button that goes to the repository.

Nothing is hidden in a `node_modules` forest. The HTML, the styles, and the
components are written in the chapters that follow, then extracted to
`index.html` and `src/`. Serve that folder with any static file server and the
app runs.

The entry script is <<main-js>>. It hangs <<app-js>> on the page. The form is
<<add-todo-component>>; each row is <<todo-item-component>>. The empty shell
is <<index-html>>.

## What Preact is

[React](https://react.dev/) taught the web a habit: describe the screen as a
function of state, and let the library update the DOM. [Preact](https://preactjs.com/)
is that habit in a small suitcase — same idea, far less machinery, happy to
live in a single `<script type="module">`.

We do not use JSX (the HTML-looking syntax that needs a compiler). We use
[htm](https://github.com/developit/htm): tagged templates. You write
`` html`<button>Add</button>` `` and the browser already understands it. An
*import map* in <<index-html>> points the words `preact` and `htm/preact` at
CDN copies, so there is no install step to *run* the app.

If you have only written Python notebooks: think of Preact as the part that
redraws the output cell whenever the variables change. `useState` is a cell
that remembers a value between redraws — the list of todos, which filter is
on.

## How the app works

A todo is a small record: an id, a title, a done flag. <<app-js>> keeps an
array of them, plus a filter (`all`, `active`, `done`).

- <<app-add-todo>> appends a new item when the form in <<add-todo-component>>
  submits.
- <<app-toggle-todo>> flips `done` when you click the checkbox on
  <<todo-item-component>>.
- <<app-delete-todo>> removes a row.
- <<app-filter-todos>> and <<app-todo-ul>> decide what you see.

The amber and black, the hard shadows, the quiet type — that is
<<style-css>>, drawn after the look of [lit-lang.org](https://lit-lang.org/).
The page should feel like paper, not a dashboard.

You can skip the next sections if you only wanted to understand the idea.
They are for when you want to cook the book yourself.

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
| Language inference | `.js` → JavaScript, `.css` → CSS, `.json` → JSON in rendered output |
| `goweb tangle` | Extract all chunks into runnable source files |
| `goweb render` | Generate full HTML documentation with the light theme |
| `goweb weave` | Produce clean markdown with fenced code blocks |
| Chunk order independence | Handler functions defined after their usage in the render tree |
| Cross-file chunk references | Components imported from `components.md`, config from `config.md` |

## Usage

```bash
# Tangle all source files into the examples/preact-todo/ directory
goweb tangle --var APP_NAME="Todo App" main.md

# With debug logging (adds console.log statements)
goweb tangle --var debug=true main.md

# Serve as static files (required for ES modules)
# Any static server works; .js is sent as JavaScript, unlike .tsx
python -m http.server 8080
# then open http://localhost:8080/

# Generate HTML documentation with the light theme
goweb render main.md --var debug=true --var APP_NAME="Todo App" -o docs.html
```

## How the chunks are organized

The example is split across four files (open them as markdown if you prefer the raw book):

- **[main.md](main.md)** (this file) — entry point with imports, overview, and the LICENSE chunk
- **[app.md](app.md)** — the room: state, filters, and how the page is hung
- **[components.md](components.md)** — the three gestures: add, list, row
- **[config.md](config.md)** — the HTML envelope and a small `package.json`

Each file exports named chunks that are resolved during tangling. The `<<import "file.md">>>` directive merges the imported file's chunks into the namespace, so chunks defined in `components.md` can be referenced from `app.md`.

## Chunk overriding

The `TodoList` component chunk is tagged with `override-demo`. If you define a chunk with the same name
in another file and mark it with `<<override>>`, goweb will replace the original definition.
This is useful for swapping implementations without modifying the original source:

```bash
# Create an override chunk
echo '<<todo-list-component>>= file: src/components/todo-list.js
export function TodoList() { return null; }
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
├── package.json        # Tangled metadata
├── index.html          # HTML + import map
└── src/
    ├── main.js         # Preact mount (browser ES module)
    ├── app.js          # App component
    ├── style.css       # lit-lang.org-inspired styles
    └── components/
        ├── add-todo.js
        ├── todo-list.js
        └── todo-item.js
```

## Conditional compilation with <<if>>

The `app.md` file uses `<<if debug>>` / `<<end>>` directives to conditionally include
`console.log` statements. When `--var debug=true` is passed, the debug chunks are included;
otherwise they are stripped. This allows a single source to produce both production
and development builds.

## Tags, graphs, and talking to a model

A literate program is a pile of named pieces. After a while the pile is
large enough that *you* still know the plot, but a new reader — human or
model — does not. Two tools keep the plot visible: **tags** on the chunks,
and a **graph** of who points at whom.

### Tags are shelves, not comments

A tag is a word you put on a chunk header:

```
<<add-todo-component>>= file: src/components/add-todo.js tags: component
```

It does not change the running app. It is a label for later. In this book
we use four shelves:

- **`component`** — anything that draws: <<add-todo-component>>,
  <<todo-item-component>>, <<todo-list-component>>, <<app-js>>
- **`config`** — the envelope: <<index-html>>, <<package-json>>
- **`debug`** — the `console.log` asides, only present when `--var debug=true`
- **`meta`** — the license and other paperwork

`goweb tangle --match component main.md` writes *only* the chunks on that
shelf. You can keep a full program in one tree and still extract “just the
UI” or “just the config” the way you would pull a chapter, not the whole
volume.

### The graph is the table of contents for dependencies

`goweb index main.md` prints every chunk, where it was defined, and who
mentions it. `goweb graph main.md` prints the same relationships as
[Graphviz](https://graphviz.org/) DOT — boxes and arrows:

```bash
goweb index main.md
goweb graph main.md | dot -Tsvg -o deps.svg
goweb graph --cluster main.md | dot -Tpng -o deps.png
```

`--cluster` groups chunks by the file they tangle into, so you see
*neighborhoods* (everything that becomes `src/app.js`) instead of a hairball.
An arrow `A → B` means A’s body contains `<<B>>`. That is the same
cross-reference the HTML uses when you click a chunk name on this page.

If two chunks should talk and the graph shows no edge, the book is lying or
unfinished. If the graph has a cycle, tangle will refuse — the story cannot
be put in order.

### Why this helps an LLM

A language model does not wander a repo the way you do. It sees what you
put in the prompt. A literate program plus tags and a graph is a polite
way to pack that prompt:

1. **Start with the map, not the files.** Paste `goweb index` or `goweb graph`
   first. The model learns the cast of characters (`<<app-js>>` uses
   `<<app-add-todo>>`) without eating every line of JavaScript.
2. **Fetch by shelf.** “Here are all `component` chunks” is a smaller,
   cleaner context than “here is `src/`.” Use `--match` to tangle or copy
   only that shelf into the conversation.
3. **Follow one edge.** To change the add form, you do not dump the
   program. You give <<add-todo-component>> and the one node that calls
   it (<<app-js>>). The graph tells you there is nothing else.
4. **Keep the prose.** The sentences around a chunk are already the
   rationale a model otherwise has to invent. Weaving or rendering this
   book *is* documentation context — not a second artifact you forgot to
   write.
5. **Name the task in tags.** A future `tags: llm, retrieval` or
   `tags: api` shelf can mark the pieces you always want in a system
   prompt. The program becomes its own retrieval index.

In short: tags say *what kind of thing* a chunk is. The graph says *what
it is for*. Together they let you hand a model a chapter instead of a
warehouse.

## License

<<license>>= file: LICENSE tags: meta
{{APP_NAME}}
Copyright (c) 2026 {{AUTHOR}} <{{EMAIL}}>

This work is licensed under the Creative Commons Attribution-ShareAlike
4.0 International License.

https://creativecommons.org/licenses/by-sa/4.0/
>>
