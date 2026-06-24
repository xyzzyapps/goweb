# UI Components

This file defines reusable Preact components for the TODO app.
Each component is a chunk with a `file:` attribute for tangling.

<<add-todo-component>>= file: src/components/add-todo.tsx tags: component
import { useState } from "preact/hooks";

interface AddTodoProps {
  onAdd: (title: string) => void;
}

export function AddTodo({ onAdd }: AddTodoProps) {
  const [value, setValue] = useState("");

  function handleSubmit(e: Event) {
    e.preventDefault();
    const trimmed = value.trim();
    if (trimmed === "") return;
    onAdd(trimmed);
    setValue("");
  }

  return (
    <form onSubmit={handleSubmit} class="flex gap-2">
      <input
        type="text"
        value={value}
        onInput={(e) => setValue((e.target as HTMLInputElement).value)}
        placeholder="What needs to be done?"
        class="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-400"
      />
      <button
        type="submit"
        class="px-4 py-2 bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition-colors"
      >
        Add
      </button>
    </form>
  );
}
>>

<<todo-list-component>>= file: src/components/todo-list.tsx tags: component, override-demo
// This component is imported by app.tsx inline via Array.map.
// It exists as a standalone export for potential reuse.
import type { JSX } from "preact";

interface Todo {
  id: number;
  title: string;
  done: boolean;
}

interface TodoListProps {
  todos: Todo[];
  onToggle: (id: number) => void;
  onDelete: (id: number) => void;
}

// In a more complex app this would handle virtualization or grouping.
export function TodoList({ todos, onToggle, onDelete }: TodoListProps) {
  return (
    <ul class="space-y-2">
      {todos.map((todo) => (
        <li key={todo.id}>
          <TodoItem
            todo={todo}
            onToggle={onToggle}
            onDelete={onDelete}
          />
        </li>
      ))}
    </ul>
  );
}
>>

<<todo-item-component>>= file: src/components/todo-item.tsx tags: component
import type { JSX } from "preact";

interface Todo {
  id: number;
  title: string;
  done: boolean;
}

interface TodoItemProps {
  todo: Todo;
  onToggle: (id: number) => void;
  onDelete: (id: number) => void;
}

export function TodoItem({ todo, onToggle, onDelete }: TodoItemProps) {
  return (
    <div class="flex items-center gap-3 p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors group">
      <input
        type="checkbox"
        checked={todo.done}
        onClick={() => onToggle(todo.id)}
        class="w-5 h-5 text-blue-500 rounded focus:ring-blue-400 cursor-pointer"
      />
      <span
        class={"flex-1 " + (todo.done ? "line-through text-gray-400" : "text-gray-800")}
      >
        {todo.title}
      </span>
      <button
        onClick={() => onDelete(todo.id)}
        class="opacity-0 group-hover:opacity-100 px-2 py-1 text-sm text-red-500 hover:bg-red-100 rounded transition-all"
        aria-label="Delete"
      >
        ✕
      </button>
    </div>
  );
}
>>
