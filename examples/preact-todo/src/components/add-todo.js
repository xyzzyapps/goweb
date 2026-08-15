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
