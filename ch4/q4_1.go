package main

import (
	"flag"
	"time"
)

func main() {
	d := flag.Int64("d", 1, "duration")
	flag.Parse()
	<-time.After(time.Duration(*d) * time.Second)
}
