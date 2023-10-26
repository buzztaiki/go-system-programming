package main

import "fmt"

// go run -gcflags -m ./ch16_1_5_stack_or_heap.go

func main() {
	a := [4]int{1, 2, 3, 4}
	b := make([]int, 4)
	c := make([]int, 4, 16)
	d := make(map[string]int)
	e := make(map[string]int, 100)
	f := make(chan string)
	g := make(chan string, 10)

	fmt.Println(a, b, c, d, e, f, g)

	aa := [4]int{1, 2, 3, 4}
	for _, xaa := range aa {
		fmt.Println(xaa)
	}

	bb := make([]int, 4)
	for _, xbb := range bb {
		fmt.Println(xbb)
	}
}
