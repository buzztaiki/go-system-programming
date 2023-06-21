package main

import (
	"archive/zip"
	"io"
	"os"
	"strings"
)

// archive/zip パッケージを使ってzip ファイルを作成してみましょう。
func main() {
	w, err := os.Create("a.zip")
	if err != nil {
		panic(err)
	}
	defer w.Close()

	zw := zip.NewWriter(w)
	defer zw.Close()

	entry, err := zw.Create("newfile.txt")
	if err != nil {
		panic(err)
	}
	io.Copy(entry, strings.NewReader("xxxx"))
	io.WriteString(w io.Writer, s string)
}
