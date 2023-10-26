package main

import "fmt"

func main() {
	h1 := map[string]int{"apple": 100, "banana": 200}
	fmt.Println(h1["apple"])
	h1["apple"] = 300
	fmt.Println(h1["apple"])

	type Key struct {
		name  string
		sound string
	}

	h2 := map[Key]int{{"cow", "moo"}: 1000, {"cat", "mew"}: 2000}
	fmt.Println(h2[Key{"cow", "moo"}])
	h2[Key{"cow", "moo"}] = 3000
	fmt.Println(h2[Key{"cow", "moo"}])

	fmt.Println(Key{"cow", "moo"} == Key{"cow", "moo"})
}
