// Custom implementation here
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
