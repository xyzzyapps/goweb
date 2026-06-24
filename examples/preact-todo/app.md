# App Component

The App component is the main entry point of the Preact application.
It manages global state (the todo list and active filter) using Preact's `useState` hook,
and orchestrates the three sub-components defined in `components.md`:
`AddTodo`, `TodoList`, and `TodoItem`.

The file exports two chunks:
- `src/style.css` — Tailwind CSS directives for base styling
- `src/main.tsx` — Mounts the `<App />` component to the DOM
- `src/app.tsx` — The full App component with state management and filter UI

Each handler function (add, toggle, delete) is defined as a named chunk
to demonstrate goweb's out-of-order chunk definitions. The debug logging
chunks (`<<debug-log-*>>`) are only included when `--var debug=true`.

<<style-css>>= file: src/style.css tags: component
@tailwind base;
@tailwind components;
@tailwind utilities;

body {
  font-family: system-ui, -apple-system, "Segoe UI", Roboto, sans-serif;
}
>>

<<main-entry>>= file: src/main.tsx tags: component
import { render } from "preact";
import { App } from "./app";
import "./style.css";

const root = document.getElementById("app");
if (root) {
  render(<App />, root);
}
>>

<<app-component>>= file: src/app.tsx tags: component
import { useState } from "preact/hooks";
import { AddTodo } from "./components/add-todo";
import { TodoList } from "./components/todo-list";
import { TodoItem } from "./components/todo-item";

interface Todo {
  id: number;
  title: string;
  done: boolean;
}

export function App() {
  const [todos, setTodos] = useState<Todo[]>([]);
  const [filter, setFilter] = useState<"all" | "active" | "done">("all");

  <<app-add-todo>>
  <<app-toggle-todo>>
  <<app-delete-todo>>
  <<app-filter-todos>>

  return (
    <div class="max-w-lg mx-auto mt-10 p-6 bg-white rounded-xl shadow-lg">
      <h1 class="text-3xl font-bold text-gray-800 mb-6 text-center">
        {{APP_NAME}}
      </h1>

      <AddTodo onAdd={handleAdd} />

      <div class="flex gap-2 my-4">
        <button
          onClick={() => setFilter("all")}
          class={"px-3 py-1 rounded-full text-sm " +
            (filter === "all"
              ? "bg-blue-500 text-white"
              : "bg-gray-200 text-gray-700")
          }
        >
          All
        </button>
        <button
          onClick={() => setFilter("active")}
          class={"px-3 py-1 rounded-full text-sm " +
            (filter === "active"
              ? "bg-blue-500 text-white"
              : "bg-gray-200 text-gray-700")
          }
        >
          Active
        </button>
        <button
          onClick={() => setFilter("done")}
          class={"px-3 py-1 rounded-full text-sm " +
            (filter === "done"
              ? "bg-blue-500 text-white"
              : "bg-gray-200 text-gray-700")
          }
        >
          Done
        </button>
      </div>

      <<app-todo-ul>>

      <div class="mt-4 text-sm text-gray-500 text-center">
        {todos.length} item{todos.length !== 1 ? "s" : ""} total
      </div>
    </div>
  );
}
>>

<<app-add-todo>>=
function handleAdd(title: string) {
  const newTodo: Todo = {
    id: Date.now(),
    title,
    done: false,
  };
  setTodos((prev) => [...prev, newTodo]);
  <<debug-log-add>>
}
>>

<<app-toggle-todo>>=
function handleToggle(id: number) {
  setTodos((prev) =>
    prev.map((t) => (t.id === id ? { ...t, done: !t.done } : t))
  );
  <<debug-log-toggle>>
}
>>

<<app-delete-todo>>=
function handleDelete(id: number) {
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
<ul class="space-y-2">
  {filtered.map((todo) => (
    <li key={todo.id}>
      <TodoItem
        todo={todo}
        onToggle={handleToggle}
        onDelete={handleDelete}
      />
    </li>
  ))}
  {filtered.length === 0 && (
    <li class="text-gray-400 text-center py-4">
      No todos yet. Add one above!
    </li>
  )}
</ul>
>>

<<debug-log-add>>= tags: debug
<<if debug>>
console.log("[debug] added todo:", title);
<<end>>
>>

<<debug-log-toggle>>= tags: debug
<<if debug>>
console.log("[debug] toggled todo:", id);
<<end>>
>>

<<debug-log-delete>>= tags: debug
<<if debug>>
console.log("[debug] deleted todo:", id);
<<end>>
>>
