// Q3.4：zip ファイルをウェブサーバーからダウンロード

package main

import (
	"archive/zip"
	"io"
	"net/http"
	"strings"
)

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=ascii_sample.zip")
	zw := zip.NewWriter(w)
	defer zw.Close()

	entry, err := zw.Create("newfile.txt")
	if err != nil {
		panic(err)
	}
	io.Copy(entry, strings.NewReader("xxxx"))
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
