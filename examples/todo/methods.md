# Methods on Todos

Notice: `<<types>>` is defined in `types.md` — cross-file reference.

<<methods>>=
<<method-list>>
<<method-add>>
<<method-done>>
>>

<<method-list>>=
func (t Todos) List() {
	for _, todo := range t {
		mark := " "
		if todo.Done {
			mark = "x"
		}
		fmt.Printf("[%s] %d: %s\n", mark, todo.Id, todo.Title)
	}
}
>>

<<method-add>>=
func (t *Todos) Add(title string) {
	id := len(*t) + 1
	*t = append(*t, Todo{Id: id, Title: title, Done: false})
}
>>

<<method-done>>=
func (t *Todos) Done(id int) error {
	for i, todo := range *t {
		if todo.Id == id {
			(*t)[i].Done = true
			return nil
		}
	}
	return fmt.Errorf("todo %d not found", id)
}
>>
