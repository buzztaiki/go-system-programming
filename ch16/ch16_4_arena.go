package main

// export GOEXPERIMENT=arenas
// go run ./ch16_4_arena.go

import (
	"arena"
	"fmt"
)

type Book struct {
	title string
}

func main() {
	a := arena.NewArena()   // アリーナ作成
	b := arena.New[Book](a) // アリーナ内に Book 型のメモリを作成
	*b = Book{"The Go Programming Language"}
	s := arena.MakeSlice[Book](a, 10, 10) // アリーナ内に []Book 型のスライス作成
	b2 := arena.Clone(b)                  // アリーナ内のメモリをヒープにコピー

	fmt.Println("arena", a)
	fmt.Println("values", b, b2, s)

	a.Free()

	fmt.Println("arena after free", a)
	// CC=clang go run -asan ./ch16_4_arena.go で実行するとエラーになる
	// see https://zenn.dev/koya_iwamura/articles/caa9cd286c4734
	fmt.Println("values after free", b, b2, s)

	// b2 はヒープにあるから -asan してもエラーにならない
	fmt.Println("b2 after free", b2)
}
