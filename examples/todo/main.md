# TODO App — Main Entry Point

<<import "types.md">>
<<import "methods.md">>

This file imports chunks from `types.md` and `methods.md`.
All chunks across all files are merged into a single namespace,
so order doesn't matter — chunks can be defined anywhere and
referenced from anywhere.

<<package>>= file: main.go
package main

import (
	"fmt"
	"os"
	"strings"

	<<imports>>
)

<<types>>
<<methods>>
<<main>>
>>

<<imports>>=
<<if debug>>
	"log"
<<end>>
>>

<<main>>=
func main() {
	todos := Todos{}
	<<init-todos>>
	<<debug-init>>
	<<run-loop>>
}
>>

<<license>>= file: LICENSE
MIT License

Copyright (c) 2026 goweb
>>

<<debug-init>>=
<<if debug>>
	log.Printf("initialized %d todos", len(todos))
<<end>>
>>

<<init-todos>>=
todos = append(todos,
	Todo{Id: 1, Title: "Learn goweb", Done: false},
	Todo{Id: 2, Title: "Write docs", Done: false},
)
>>

<<run-loop>>=
for {
	fmt.Print("> ")
	var input string
	fmt.Scanln(&input)
	parts := strings.Fields(input)
	if len(parts) == 0 {
		continue
	}
	switch parts[0] {
	case "list":
		todos.List()
	case "add":
		<<handle-add>>
	case "done":
		<<handle-done>>
	case "exit":
		<<handle-exit>>
	default:
		fmt.Println("commands: list, add <title>, done <id>, exit")
	}
}
>>

<<handle-add>>=
if len(parts) < 2 {
	fmt.Println("usage: add <title>")
	continue
}
title := strings.Join(parts[1:], " ")
todos.Add(title)
<<debug-add>>
>>

<<debug-done>>=
<<if debug>>
	log.Printf("marked todo %d as done", id)
<<end>>
>>

<<handle-exit>>=
<<if debug>>
	log.Println("shutting down")
<<end>>
	os.Exit(0)
>>

<<debug-add>>=
<<if debug>>
	log.Printf("added todo: %s", title)
<<end>>
>>

<<handle-done>>=
if len(parts) < 2 {
	fmt.Println("usage: done <id>")
	continue
}
var id int
fmt.Sscanf(parts[1], "%d", &id)
if err := todos.Done(id); err != nil {
	fmt.Println("error:", err)
}
<<debug-done>>
>>
