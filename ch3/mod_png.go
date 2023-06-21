package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"os"
)

func dumpChunk(chunk io.Reader) {
	var length int32
	binary.Read(chunk, binary.BigEndian, &length)
	buffer := make([]byte, 4)
	chunk.Read(buffer)
	fmt.Printf("chunk '%v' (%d bytes)\n", string(buffer), length)
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
		// SectionReader はランダムアクセスだからポインタは移動しない
		chunks = append(chunks,
			io.NewSectionReader(file, offset, int64(length)+12))
		// 次のチャンクの先頭に移動
		// 現在位置は長さを読み終わった箇所なので
		// チャンク名(4 バイト) + データ長+ CRC(4 バイト) 先に移動
		offset, _ = file.Seek(int64(length+8), 1)
	}
	return chunks
}
func textChunk(text string) io.Reader {
	byteText := []byte(text)
	crc := crc32.NewIEEE()
	var buffer bytes.Buffer
	binary.Write(&buffer, binary.BigEndian, int32(len(byteText)))
	// CRC 計算とバッファへの書き込みを同時に行うMultiWriter
	writer := io.MultiWriter(&buffer, crc)
	io.WriteString(writer, "teXt") // 2 バイトめの5 ビットめを立てる( 小文字にする) とプライベート
	writer.Write(byteText)
	binary.Write(&buffer, binary.BigEndian, crc.Sum32())
	return &buffer
}
func main() {
	file, err := os.Open("PNG_transparency_demonstration_1.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	newFile, err := os.Create("PNG_transparency_demonstration_secret.png")
	if err != nil {
		panic(err)
	}
	defer newFile.Close()
	chunks := readChunks(file)
	// シグニチャ書き込み
	io.WriteString(newFile, "\x89PNG\r\n\x1a\n")
	// 先頭に必要なIHDR チャンクを書き込み
	io.Copy(newFile, chunks[0])
	// テキストチャンクを追加
	io.Copy(newFile, textChunk("Lambda Note++"))
	// 残りのチャンクを追加
	for _, chunk := range chunks[1:] {
		io.Copy(newFile, chunk)
	}
}
