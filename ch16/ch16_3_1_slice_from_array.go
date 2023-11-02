package main

import "fmt"

func main() {
	a := [4]int{1, 2, 3, 4}
	// 既存の配列を参照するスライス
	b := a[:]
	fmt.Println(&a[0], &b[0], len(b), cap(b))

	// 既存の配列の一部を参照するスライス
	c := a[1:3]
	fmt.Println(&a[1], &c[0], len(c), cap(c))

	fmt.Printf("&a=%p, &b=%p, &c=%p\n", &a, &b, &c)

	a0 := a[0]
	fmt.Printf("&a0=%p, &a[0]=%p\n", &a0, &a[0])

	b[1] = 10
	fmt.Println(a)
	fmt.Println(b)
	fmt.Println(c)
}
