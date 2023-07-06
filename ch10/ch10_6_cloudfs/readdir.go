package main

import (
	"context"
	"strings"

	"github.com/billziss-gh/cgofuse/fuse"
	"gocloud.dev/blob"
)

func (cf *CloudFileSystem) Readdir(
	path string,
	// ディレクトリエントリを埋める用のコールバックが渡ってくる
	fill func(name string, stat *fuse.Stat_t, ofst int64) bool,
	ofst int64,
	fh uint64,
) int {
	ctx := context.Background()

	// 必ず必要なやつを埋める
	fill(".", nil, 0)
	fill("..", nil, 0)

	// クラウド用のディレクトリ名にする
	prefix := strings.TrimLeft(path, "/")
	if prefix != "" {
		prefix = prefix + "/"
	}

	i := cf.bucket.List(&blob.ListOptions{
		Prefix:    prefix,
		Delimiter: "/",
	})

	for {
		// イテレータが自然に扱えないのが悲しい
		o, err := i.Next(ctx)
		if err != nil {
			break
		}

		// ディレクトリ名を除外
		key := o.Key[len(prefix):]
		if len(key) == 0 {
			continue
		}

		// ディレクトリ末尾の / を除外して埋める
		fill(strings.TrimRight(key, "/"), nil, 0)
	}

	// 成功
	return 0
}
