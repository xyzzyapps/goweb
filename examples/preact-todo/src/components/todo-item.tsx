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
