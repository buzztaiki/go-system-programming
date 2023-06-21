package main

import (
	"context"
	"fmt"
	"math"
	"time"
)

func primeNumber(ctx context.Context) chan int {
	result := make(chan int)
	go func() {
		result <- 2
	loop:
		for i := 3; i < 100000; i += 2 {
			l := int(math.Sqrt(float64(i)))
			found := false
			for j := 3; j < l+1; j += 2 {
				select {
				case <-ctx.Done():
					fmt.Println("done")
					break loop
				default:
					if i%j == 0 {
						found = true
						break
					}
				}

			}
			if !found {
				result <- i
			}
		}
		close(result)
	}()
	return result
}

func main() {
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Millisecond)
	pn := primeNumber(ctx)
	for n := range pn {
		fmt.Println(n)
	}
}
