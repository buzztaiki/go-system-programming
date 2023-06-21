package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

func dumpChunk(chunk io.Reader) {
	var length int32
	binary.Read(chunk, binary.BigEndian, &length)
	buffer := make([]byte, 4)
	chunk.Read(buffer)
	fmt.Printf("chunk '%v' (%d bytes)\n", string(buffer), length)
	if bytes.Equal(buffer, []byte("teXt")) {
		rawText := make([]byte, length)
		chunk.Read(rawText)
		fmt.Println(string(rawText))
	}
}
func readChunks(file *os.File) []io.Reader {
	// チャンクを格納する配列
	var chunks []io.Reader
	// 最初の8 バイトを飛ばす
	file.Seek(8, 0)
	var offset int64 = 8
	for {
		var length int32
		err := binary.Read(file, binary.BigEndian, &length)
		if err == io.EOF {
			break
		}
		// SectionReader を作って chunks に追加してるだけ
		chunks = append(chunks,
			io.NewSectionReader(file, offset, int64(length)+12))
		// 次のチャンクの先頭に移動
		// 現在位置は長さを読み終わった箇所なので
		// チャンク名(4 バイト) + データ長+ CRC(4 バイト) 先に移動
		offset, _ = file.Seek(int64(length+8), 1)
	}
	return chunks
}
func main() {
	file, err := os.Open("PNG_transparency_demonstration_1.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	chunks := readChunks(file)
	for _, chunk := range chunks {
		dumpChunk(chunk)
	}
}
