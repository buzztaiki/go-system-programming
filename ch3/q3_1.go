package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

// 古いファイル（old.txt）を新しいファイル（new.txt）にコピーしてみましょう。
// 本章で紹介したサンプルコードを応用すれば難しくないと思います。
// さらに改造して実用的なコマンドにしてみたいと思われる方は、コマンドラインオ
// プションでファイル名を渡せるようにするとよいでしょう。本書の範囲からは外れ
// るので詳細は省きますが、os.Args という文字列配列にオプションが格納されます。
// また、標準ライブラリにあるflag パッケージを使うと、オプションのパース処理が
// より便利に行えます。
func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "usage: %s <src> <dst>:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	if len(flag.Args()) != 2 {
		flag.Usage()
		os.Exit(2)
	}

	src, err := os.Open(flag.Arg(0))
	if err != nil {
		log.Fatalf("failed to open src file: %v", err)
	}
	defer src.Close()

	dst, err := os.Create(flag.Arg(1))
	if err != nil {
		log.Fatalf("failed to open dest file: %v", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		log.Fatalf("failed to copy from src to dest: %v", err)
	}
}
