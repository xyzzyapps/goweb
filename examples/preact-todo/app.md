# The page you actually use

This chapter is the app's front door. <<app-js>> is the room: it remembers
the list and the current filter, and it asks the smaller pieces — add, row,
list — to draw themselves. <<main-js>> only does one job: find the empty
`<div id="app">` in <<index-html>> and place that room inside it.

If you have used a notebook, this file is the cell that *owns the variables*.
The hooks below are those variables. The markup below is the output.

## Chunks in this file

This file exports these files when tangled:

- `src/style.css` — lit-lang.org-inspired styles
- `src/main.js` — Mounts the `App` component (browser ES module)
- `src/app.js` — The full App component with state management and filter UI

## Goweb features shown

**Chunk order independence.** The handler functions (`handleAdd`, `handleToggle`, `handleDelete`)
are defined as separate named chunks (`<<app-add-todo>>`, `<<app-toggle-todo>>`, `<<app-delete-todo>>`)
and referenced by `<<name>>` inside the `App` component. This means they can be defined
**after** the render function that uses them — goweb resolves all references during tangling,
so definition order doesn't matter.

**Conditional compilation.** Debug logging chunks (`<<debug-log-add>>`, `<<debug-log-toggle>>`,
`<<debug-log-delete>>`) are wrapped in `<<if debug>>` / `<<end>>` directives. When you tangle
with `--var debug=true`, these chunks are included. When `debug` is not set, they are stripped.
This lets you maintain debug code alongside production code without runtime overhead.

**Cross-file references.** The `app.js` chunk imports from `./components/add-todo.js`,
`./components/todo-list.js`, and `./components/todo-item.js`. These components are defined
in `components.md` and imported via `<<import "components.md">>` in `main.md`.

**Chunk references for stubs.** The `app-todo-ul` reference is left as a placeholder
inside the JSX. During tangling, any chunk named `<<app-todo-ul>>` (defined elsewhere)
will be substituted in its place. This is useful for out-of-order definition.

<<style-css>>= file: src/style.css tags: component
/* Palette and chrome match https://lit-lang.org/ */
:root {
  --color-primary: #ffcd42;
  --color-primary-light: rgba(255, 205, 66, 0.12);
  --color-primary-darker: #291f04;
  --color-primary-text: #333;
  --gradient-primary: linear-gradient(to right, #ffcd42, #ff6102);
}

*, *::before, *::after { box-sizing: border-box; }

body {
  margin: 0;
  min-height: 100vh;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Noto Sans", Helvetica, Arial, sans-serif;
  font-size: 18px;
  line-height: 1.5;
  color: var(--color-primary-text);
  background: #fff;
}

.site-header {
  padding: 1rem 1rem 0.25rem;
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem 1.5rem;
  max-width: 800px;
  margin: 0 auto;
}

.site-brand {
  font-weight: 700;
  font-size: 1.5rem;
  color: inherit;
  text-decoration: none;
}

.site-nav {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.6rem;
}

.view-source, .github-link {
  font-weight: 600;
  color: #000;
  text-decoration: underline;
  padding: 0.4rem 0.85rem;
  border: 2px solid #000;
  background: var(--color-primary);
  box-shadow: 3px 3px 0 #000;
}

.github-link { background: #fff; }

.view-source:hover, .github-link:hover { text-decoration: none; }

.todo-app {
  max-width: 800px;
  margin: 0 auto;
  padding: 1rem 1rem 3rem;
}

.todo-app h1 {
  margin: 0 0 1.25rem;
  padding-bottom: 0.5rem;
  background: linear-gradient(transparent, transparent) no-repeat 0 0,
    var(--gradient-primary) no-repeat 0 calc(100% - 0.5px) / 100% 2px;
}

.todo-form { display: flex; gap: 0.5rem; }

.todo-input, .todo-btn, .filter-btn, .todo-delete {
  font-family: inherit;
  font-size: 1rem;
}

.todo-input {
  flex: 1;
  padding: 0.55rem 0.75rem;
  border: 2px solid #000;
  background: #fff;
  box-shadow: 3px 3px 0 #000;
}

.todo-btn, .filter-btn {
  padding: 0.55rem 0.9rem;
  border: 2px solid #000;
  background: var(--color-primary);
  font-weight: 600;
  cursor: pointer;
  box-shadow: 3px 3px 0 #000;
}

.filter-btn.is-active {
  background: #ff6102;
  color: #fff;
}

.filters { display: flex; gap: 0.5rem; margin: 1rem 0; flex-wrap: wrap; }

.todo-list { list-style: none; padding: 0; margin: 0; }

.todo-row {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.7rem 0.85rem;
  margin-bottom: 0.6rem;
  border: 2px solid #000;
  background: var(--color-primary-light);
  box-shadow: 3px 3px 0 #000;
}

.todo-title.is-done { text-decoration: line-through; opacity: 0.55; }
.todo-title { flex: 1; }

.todo-delete {
  border: 2px solid #000;
  background: #fff;
  cursor: pointer;
  padding: 0.15rem 0.5rem;
  font-weight: 700;
}

.todo-empty, .todo-count { text-align: center; color: #666; margin-top: 1rem; }

.site-footer {
  background: var(--color-primary-darker);
  color: #fff;
  text-align: center;
  padding: 1.5rem 1rem;
  font-size: 0.9rem;
}
.site-footer a { color: var(--color-primary); font-weight: 600; }
>>

<<main-js>>= file: src/main.js tags: component
import { render } from "preact";
import { html } from "htm/preact";
import { App } from "./app.js";

const root = document.getElementById("app");
if (root) {
  render(html`<${App} />`, root);
}
>>

<<app-js>>= file: src/app.js tags: component
import { useState } from "preact/hooks";
import { html } from "htm/preact";
import { AddTodo } from "./components/add-todo.js";
import { TodoItem } from "./components/todo-item.js";

export function App() {
  const [todos, setTodos] = useState([]);
  const [filter, setFilter] = useState("all");

  <<app-add-todo>>
  <<app-toggle-todo>>
  <<app-delete-todo>>
  <<app-filter-todos>>

  return html`
    <header class="site-header">
      <a class="site-brand" href="#">goweb</a>
      <nav class="site-nav">
        <a class="view-source" href="docs.html">View source</a>
        <a class="github-link" href="{{REPO}}">GitHub</a>
      </nav>
    </header>
    <div class="todo-app">
      <h1>{{APP_NAME}}</h1>
      <${AddTodo} onAdd=${handleAdd} />
      <div class="filters">
        <button
          onClick=${() => setFilter("all")}
          class=${"filter-btn" + (filter === "all" ? " is-active" : "")}
        >
          All
        </button>
        <button
          onClick=${() => setFilter("active")}
          class=${"filter-btn" + (filter === "active" ? " is-active" : "")}
        >
          Active
        </button>
        <button
          onClick=${() => setFilter("done")}
          class=${"filter-btn" + (filter === "done" ? " is-active" : "")}
        >
          Done
        </button>
      </div>
      <<app-todo-ul>>
      <div class="todo-count">
        ${todos.length} item${todos.length !== 1 ? "s" : ""} total
      </div>
    </div>
  `;
}
>>

<<app-add-todo>>=
function handleAdd(title) {
  const newTodo = {
    id: Date.now(),
    title,
    done: false,
  };
  setTodos((prev) => [...prev, newTodo]);
  <<debug-log-add>>
}
>>

<<app-toggle-todo>>=
function handleToggle(id) {
  setTodos((prev) =>
    prev.map((t) => (t.id === id ? { ...t, done: !t.done } : t))
  );
  <<debug-log-toggle>>
}
>>

<<app-delete-todo>>=
function handleDelete(id) {
  setTodos((prev) => prev.filter((t) => t.id !== id));
  <<debug-log-delete>>
}
>>

<<app-filter-todos>>=
const filtered = todos.filter((t) => {
  if (filter === "active") return !t.done;
  if (filter === "done") return t.done;
  return true;
});
>>

<<app-todo-ul>>=
<ul class="todo-list">
  ${filtered.map((todo) => html`
    <li key=${todo.id}>
      <${TodoItem}
        todo=${todo}
        onToggle=${handleToggle}
        onDelete=${handleDelete}
      />
    </li>
  `)}
  ${filtered.length === 0 && html`
    <li class="todo-empty">No todos yet. Add one above!</li>
  `}
</ul>
>>

<<if debug>>
<<debug-log-add>>= tags: debug
console.log("[debug] added todo:", title);
>>

<<debug-log-toggle>>= tags: debug
console.log("[debug] toggled todo:", id);
>>

<<debug-log-delete>>= tags: debug
console.log("[debug] deleted todo:", id);
>>
<<end>>
