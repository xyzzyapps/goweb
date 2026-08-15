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
