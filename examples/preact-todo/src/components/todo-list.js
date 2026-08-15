export function TodoList() { return null; }
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
