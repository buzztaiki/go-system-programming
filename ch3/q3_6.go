package main

import (
	"io"
	"os"
	"strings"
)

var (
	computer    = strings.NewReader("COMPUTER")
	system      = strings.NewReader("SYSTEM")
	programming = strings.NewReader("PROGRAMMING")
)

func main() {
	var stream io.Reader
	stream, w := io.Pipe()

	go func() {
		defer w.Close()
		// ASCII
		io.Copy(w, io.NewSectionReader(programming, 5, 1))
		io.Copy(w, io.NewSectionReader(system, 0, 1))
		io.Copy(w, io.NewSectionReader(computer, 0, 1))
		io.Copy(w, io.NewSectionReader(programming, 8, 1))
		io.Copy(w, io.NewSectionReader(programming, 8, 1))
	}()

	io.Copy(os.Stdout, stream)
}
