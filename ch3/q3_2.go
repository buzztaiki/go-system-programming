package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"os"
)

// ファイルを作成してランダムな内容で埋めてみましょう。
// crypto/rand パッケージ（本来は付録Aで紹介するように暗号用の機能）をイン
// ポートすると、rand.Reader というio.Reader が使えます。このReader は、ラン
// ダムなバイトを延々と出力し続ける無限長のファイルのような動作をします。これを
// 使って、1024バイトの長さのバイナリファイルを作ってみましょう。
// ヒントですが、io.Copy() を使ってはいけません。io.Copy() はReader の終了
// まですべて愚直にコピーしようとします。そしてrand.Reader には終わりはありま
// せん。あとはわかりますよね？

func main_1(dstName string) error {
	buf := make([]byte, 1024)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return fmt.Errorf("failed to read random source: %w", err)
	}
	w, err := os.Create(dstName)
	if err != nil {
		return fmt.Errorf("failed to open dest file: %w", err)
	}
	defer w.Close()

	if _, err := w.Write(buf); err != nil {
		return fmt.Errorf("failed to write to dest file: %w", err)
	}
	return nil
}

func main_2(dstName string) error {
	w, err := os.Create(dstName)
	if err != nil {
		return fmt.Errorf("failed to open dest file: %w", err)
	}
	defer w.Close()

	if _, err := io.Copy(w, io.LimitReader(rand.Reader, 1024)); err != nil {
		return fmt.Errorf("failed to write to dest file: %w", err)
	}
	return nil
}

func main() {
	if err := main_1("random1.dat"); err != nil {
		log.Fatal("failed to run main_1: %v", err)
	}
	if err := main_2("random2.dat"); err != nil {
		log.Fatal("failed to run main_2: %v", err)
	}
}
