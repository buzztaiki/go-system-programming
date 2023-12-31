package main

import (
	"fmt"
	"github.com/edsrzf/mmap-go"
	"io/ioutil"
	"os"
	"path/filepath"
)

func main() {
	// テストデータを書き込み
	var testData = []byte("0123456789ABCDEF")
	var testPath = filepath.Join(os.TempDir(), "mmap_testdata")
	err := ioutil.WriteFile(testPath, testData, 0644)
	if err != nil {
		panic(err)
	}

	// メモリにマッピング
	// m は[]byte のエイリアスなので添字アクセス可能
	f, err := os.OpenFile(testPath, os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	m, err := mmap.Map(f, mmap.RDWR, 0)
	if err != nil {
		panic(err)
	}
	defer m.Unmap()

	// メモリ上のデータを修正して書き込む
	m[9] = 'X'
	m.Flush()

	// 読み込んでみる
	fileData, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	fmt.Printf("original: %s\n", testData)
	fmt.Printf("mmap: %s\n", m)
	fmt.Printf("file: %s\n", fileData)
}
