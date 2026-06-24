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

  function handleAdd(title: string) {
  const newTodo: Todo = {
    id: Date.now(),
    title,
    done: false,
  };
  setTodos((prev) => [...prev, newTodo]);
  console.log("[debug] added todo:", title);
}
  function handleToggle(id: number) {
  setTodos((prev) =>
    prev.map((t) => (t.id === id ? { ...t, done: !t.done } : t))
  );
  console.log("[debug] toggled todo:", id);
}
  function handleDelete(id: number) {
  setTodos((prev) => prev.filter((t) => t.id !== id));
  console.log("[debug] deleted todo:", id);
}
  const filtered = todos.filter((t) => {
  if (filter === "active") return !t.done;
  if (filter === "done") return t.done;
  return true;
});

  return (
    <div class="max-w-lg mx-auto mt-10 p-6 bg-white rounded-xl shadow-lg">
      <h1 class="text-3xl font-bold text-gray-800 mb-6 text-center">
        Todo App
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

      <div class="mt-4 text-sm text-gray-500 text-center">
        {todos.length} item{todos.length !== 1 ? "s" : ""} total
      </div>
    </div>
  );
}
