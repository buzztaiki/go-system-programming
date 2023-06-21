package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// Q3.5：CopyN

func CopyN(dest io.Writer, src io.Reader, length int) (int64, error) {
	return io.Copy(dest, io.LimitReader(src, int64(length)))
}

func main() {
	fmt.Println(CopyN(os.Stdout, bytes.NewBufferString("aaaaaaaa"), 4))
}
