package main

import "fmt"

func main() {
	// 長さ 1、確保された要素 2 のスライスを作成
	s := make([]int, 1, 2)
	fmt.Println(&s[0], s, len(s), cap(s))

	// 1 要素追加（確保された範囲内）
	s1 := append(s, 1)
	fmt.Println(&s1[0], s1, len(s1), cap(s1))
	fmt.Println(&s[0], s, len(s), cap(s))

	s = s1

	// さらに要素を追加（新しく配列が確保され直す）
	// スライスの先頭要素のアドレスも変わる
	s2 := append(s, 2)
	fmt.Println(&s2[0], s2, len(s2), cap(s2))
	fmt.Println(&s[0], s, len(s), cap(s))
}
