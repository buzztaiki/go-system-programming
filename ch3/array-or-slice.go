package main

import (
	"fmt"
)

func main() {
	// make だと slice
	xs := make([]byte, 1024)
	// 以下だと配列
	ys := [1024]byte{}
	fmt.Println(len(xs), len(ys))
	fmt.Printf("%T, %T", xs, ys)
}
