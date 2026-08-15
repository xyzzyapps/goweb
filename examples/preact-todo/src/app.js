import { useState } from "preact/hooks";
import { html } from "htm/preact";
import { AddTodo } from "./components/add-todo.js";
import { TodoItem } from "./components/todo-item.js";

export function App() {
  const [todos, setTodos] = useState([]);
  const [filter, setFilter] = useState("all");

  function handleAdd(title) {
  const newTodo = {
    id: Date.now(),
    title,
    done: false,
  };
  setTodos((prev) => [...prev, newTodo]);
  console.log("[debug] added todo:", title);
}
  function handleToggle(id) {
  setTodos((prev) =>
    prev.map((t) => (t.id === id ? { ...t, done: !t.done } : t))
  );
  console.log("[debug] toggled todo:", id);
}
  function handleDelete(id) {
  setTodos((prev) => prev.filter((t) => t.id !== id));
  console.log("[debug] deleted todo:", id);
}
  const filtered = todos.filter((t) => {
  if (filter === "active") return !t.done;
  if (filter === "done") return t.done;
  return true;
});

  return html`
    <header class="site-header">
      <a class="site-brand" href="#">goweb</a>
      <nav class="site-nav">
        <a class="view-source" href="docs.html">View source</a>
        <a class="github-link" href="https://github.com/xyzzyapps/goweb">GitHub</a>
      </nav>
    </header>
    <div class="todo-app">
      <h1>Todo App</h1>
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
      <div class="todo-count">
        ${todos.length} item${todos.length !== 1 ? "s" : ""} total
      </div>
    </div>
  `;
}
