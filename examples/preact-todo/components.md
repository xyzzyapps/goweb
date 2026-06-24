# UI Components

This file defines three reusable Preact components for the TODO app.
Each component is a named chunk with a `file:` attribute that determines
the output path during tangling.

## Components

- **`AddTodo`** (`src/components/add-todo.tsx`) — Controlled form with input and submit button. Uses `useState` for local input state. The form prevents default submission and calls the parent's `onAdd` callback.
- **`TodoList`** (`src/components/todo-list.tsx`) — Standalone list component that maps over todos. Accepts `todos`, `onToggle`, and `onDelete` props. Tagged with `override-demo` to demonstrate goweb's chunk override feature.
- **`TodoItem`** (`src/components/todo-item.tsx`) — Single todo row with checkbox, title, and delete button. Uses Tailwind hover groups to show the delete button on hover, with strikethrough styling for completed items.

## Goweb features shown

**The `file:` attribute.** Each chunk definition includes `file: src/components/add-todo.tsx`,
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
<<todo-list-component>>= file: src/components/todo-list.tsx
// Custom implementation here
>>
```

The override takes precedence during tangling, making this a powerful mechanism
for swapping implementations without modifying original source files.

**Cross-file chunk resolution.** These components are imported by `app.tsx` (defined in `app.md`)
even though they are defined in this file. Goweb resolves all chunks across imported files
during tangling, so the import order in `main.md` (`<<import "app.md">>`, `<<import "components.md">>`)
ensures all chunks are available.

**No language tag needed.** The `AddTodo` and `TodoList` chunks use TypeScript/TSX syntax
but don't include a language annotation in their fenced code block. Goweb's render step
infers the language from the `file:` extension (`.tsx` → TypeScript), so the rendered
HTML documentation gets proper syntax highlighting via highlight.js.

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
