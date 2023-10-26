package main

import "fmt"

func main() {
	// 長さ 1、確保された要素 2 のスライスを作成
	s := make([]int, 1, 2)
	fmt.Println(&s[0], len(s), cap(s))

	// 1 要素追加（確保された範囲内）
	s = append(s, 1)
	fmt.Println(&s[0], len(s), cap(s))

	// さらに要素を追加（新しく配列が確保され直す）
	// スライスの先頭要素のアドレスも変わる
	s = append(s, 2)
	fmt.Println(&s[0], len(s), cap(s))
}
