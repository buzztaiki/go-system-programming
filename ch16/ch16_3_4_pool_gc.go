package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func main() {
	var count int
	pool := sync.Pool{
		New: func() interface{} {
			count++
			return fmt.Sprintf("created: %d", count)
		},
	}

	for i := 0; i < 10; i++ {
		fmt.Println("##", i)
		// GC を呼ぶと追加された要素が消える
		pool.Put("removed 1")
		pool.Put("removed 2")
		time.Sleep(1 * time.Second)
		runtime.GC()
		fmt.Println(pool.Get())
		fmt.Println(pool.Get())
		fmt.Println()
	}
}
