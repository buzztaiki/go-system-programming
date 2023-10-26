package main

import "fmt"

func main() {
	// 初期の配列とリンクされているスライス
	e := []int{1, 2, 3, 4}
	fmt.Println(&e[0], len(e), cap(e))

	// サイズを持ったスライスを定義
	f := make([]int, 4)
	fmt.Println(&f[0], len(f), cap(f))

	// サイズと容量を持ったスライスを定義
	g := make([]int, 4, 8)
	fmt.Println(&g[0], len(g), cap(g))
}
