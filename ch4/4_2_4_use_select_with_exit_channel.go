package main

import (
	"fmt"
	"math"
)

func primeNumber() (chan int, chant int, chan struct{}) {
	result := make(chan int)
	result2 := make(chan int)
	exit := make(chan struct{}, 0)
	go func() {
		result <- 2
		for i := 3; i < 100000; i += 2 {
			l := int(math.Sqrt(float64(i)))
			found := false
			for j := 3; j < l+1; j += 2 {
				if i%j == 0 {
					found = true
					break
				}
			}
			if !found {
				result <- i
			}
		}
		close(exit)
		// exit <- struct{}{}
	}()
	return result, exit
}

func main() {
	pn, pn2, exit := primeNumber()
loop:
	for {
		select {
		case n := <-pn:
			fmt.Println("pn1", n)
		case n := <-pn2:
			fmt.Println("pn2", n)
		case <-exit:
			break loop
		}
	}
	fmt.Println("end")
}
