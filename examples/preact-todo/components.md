# The small pieces of the screen

A todo app is three gestures: write something, look at the list, mark or
discard a line. Each gesture is a component — a function that takes a few
props and returns HTML. They live here so <<app-js>> can stay a story about
*state*, not about every button.

You do not need to know Preact deeply to read them. A component is “given
these values, draw this.” When a value changes, Preact draws again.

## Components

- **`AddTodo`** (`src/components/add-todo.js`) — Controlled form with input and submit button. Uses `useState` for local input state. The form prevents default submission and calls the parent's `onAdd` callback.
- **`TodoList`** (`src/components/todo-list.js`) — Standalone list component that maps over todos. Accepts `todos`, `onToggle`, and `onDelete` props. Tagged with `override-demo` to demonstrate goweb's chunk override feature.
- **`TodoItem`** (`src/components/todo-item.js`) — Single todo row with checkbox, title, and delete button. Completed items use strikethrough styling.

## Goweb features shown

**The `file:` attribute.** Each chunk definition includes `file: src/components/add-todo.js`,
which tells goweb where to write the tangled output. The path is relative to the current
working directory where `goweb tangle` is invoked.

**Chunk tags.** The `TodoList` component includes `tags: component, override-demo`. Tags
are metadata that can be used for filtering, grouping, and documentation. The `goweb graph`
command can show relationships filtered by tag.

**Chunk overriding.** The `override-demo` tag indicates that this chunk can be replaced
via goweb's `<<override>>` mechanism. To override, create another chunk with the same
name and use `<<override>>` before the definition:

```
<<override>>
<<todo-list-component>>= file: src/components/todo-list.js
export function TodoList() { return null; }
>>
```

The override takes precedence during tangling, making this a powerful mechanism
for swapping implementations without modifying original source files.

**Cross-file chunk resolution.** These components are imported by `app.js` (defined in `app.md`)
even though they are defined in this file. Goweb resolves all chunks across imported files
during tangling, so the import order in `main.md` (`<<import "app.md">>`, `<<import "components.md">>`)
ensures all chunks are available.

Tangled output is browser ES modules (`.js` + `htm`) so a static server can serve them
as `text/javascript`. JSX/TSX cannot run that way.

<<add-todo-component>>= file: src/components/add-todo.js tags: component
import { useState } from "preact/hooks";
import { html } from "htm/preact";

export function AddTodo({ onAdd }) {
  const [value, setValue] = useState("");

  function handleSubmit(e) {
    e.preventDefault();
    const trimmed = value.trim();
    if (trimmed === "") return;
    onAdd(trimmed);
    setValue("");
  }

  return html`
    <form onSubmit=${handleSubmit} class="todo-form">
      <input
        type="text"
        value=${value}
        onInput=${(e) => setValue(e.target.value)}
        placeholder="What needs to be done?"
        class="todo-input"
      />
      <button type="submit" class="todo-btn">Add</button>
    </form>
  `;
}
>>

<<todo-list-component>>= file: src/components/todo-list.js tags: component, override-demo
import { html } from "htm/preact";
import { TodoItem } from "./todo-item.js";

export function TodoList({ todos, onToggle, onDelete }) {
  return html`
    <ul class="todo-list">
      ${todos.map((todo) => html`
        <li key=${todo.id}>
          <${TodoItem}
            todo=${todo}
            onToggle=${onToggle}
            onDelete=${onDelete}
          />
        </li>
      `)}
    </ul>
  `;
}
>>

<<todo-item-component>>= file: src/components/todo-item.js tags: component
import { html } from "htm/preact";

export function TodoItem({ todo, onToggle, onDelete }) {
  return html`
    <div class="todo-row">
      <input
        type="checkbox"
        checked=${todo.done}
        onClick=${() => onToggle(todo.id)}
      />
      <span class=${"todo-title" + (todo.done ? " is-done" : "")}>
        ${todo.title}
      </span>
      <button
        onClick=${() => onDelete(todo.id)}
        class="todo-delete"
        aria-label="Delete"
      >
        ✕
      </button>
    </div>
  `;
}
>>
