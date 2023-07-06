// https://github.com/Iceber/iouring-go/blob/main/examples/concurrent-cat/main.go
// これにコメント付けたりしたやつ

package main

import (
	"fmt"
	"os"

	"github.com/iceber/iouring-go"
)

func main() {
	if len(os.Args) <= 1 {
		fmt.Printf("Usage: %s file1 file2 ...\n", os.Args[0])
	}

	iour, err := iouring.New(10)
	if err != nil {
		panic(err)
	}
	defer iour.Close()

	// 読み込んだ結果を得るチャネル
	compCh := make(chan iouring.Result, 1)

	// コマンド引数に渡されたファイル全部を送信する
	go func() {
		for _, filename := range os.Args[1:] {
			file, err := os.Open(filename)
			if err != nil {
				panic(err)
			}

			// ファイルを全部入れられるバッファを準備してるぽい
			buffers, err := prepareBuffers(file)

			// see https://man.archlinux.org/man/io_uring_prep_readv.3
			prepRequest := iouring.Readv(int(file.Fd()), buffers).WithInfo(file.Name())
			_, err = iour.SubmitRequest(prepRequest, compCh)
			if err != nil {
				panic(err)
			}
		}
	}()

	nfiles := len(os.Args) - 1

	var prints int
	// 結果をチャネルから得る
	for result := range compCh {
		filename := result.GetRequestInfo().(string)
		if err := result.Err(); err != nil {
			fmt.Printf("read %s error: %v\n", filename, result.Err())
		}

		// バッファに全部入ってるから一度に読み切る。多分。
		// こういうのをやりたくない場合は iouring.Read を使うぽい
		// - https://github.com/Iceber/iouring-go/blob/main/examples/cp/main.go
		// - https://man.archlinux.org/man/io_uring_prep_read.3.en
		fmt.Printf("%s: \n", filename)
		for _, buffer := range result.GetRequestBuffers() {
			fmt.Printf("%s", buffer)
		}
		fmt.Println()

		// 全部読み終わったら終わり
		prints++
		if prints == nfiles {
			break
		}
	}

	fmt.Println("cat successful")
}

func prepareBuffers(file *os.File) ([][]byte, error) {
	var blockSize int64 = 32 * 1024

	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	size := stat.Size()

	blocks := int(size / blockSize)
	if size%blockSize != 0 {
		blocks++
	}

	buffers := make([][]byte, blocks)
	for i := 0; i < blocks; i++ {
		buffers[i] = make([]byte, blockSize)
	}
	if size%blockSize != 0 {
		buffers[blocks-1] = buffers[blocks-1][:size%blockSize]
	}

	return buffers, nil
}
