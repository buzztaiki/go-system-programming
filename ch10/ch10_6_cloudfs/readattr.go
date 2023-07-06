package main

import (
	"context"
	"log"
	"strings"

	"github.com/billziss-gh/cgofuse/fuse"
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"
)

func (cf *CloudFileSystem) Getattr(path string, stat *fuse.Stat_t, fh uint64) (errc int) {
	if path == "/" {
		stat.Mode = fuse.S_IFDIR | 0555
		return 0
	}
	ctx := context.Background()
	// blob のパスは `/` から始まらないから除いてあげる
	name := strings.TrimLeft(path, "/")

	log.Printf("name: %s", name)

	// blob からファイル属性を取る
	a, err := cf.bucket.Attributes(ctx, name)
	if err != nil {

		// Azure Blob はディレクトリという概念がない
		if !strings.HasPrefix(cf.bucketUrl, "azblob://") {
			// 取れなかったらディレクトリとして存在してるか見る
			if exists, _ := cf.bucket.Exists(ctx, name+"/"); !exists {
				return -fuse.ENOENT
			}
		}

		// S_IFDIR だとディレクトリ。0555 はディレクトリだから実行ビット立ててる
		stat.Mode = fuse.S_IFDIR | 0555
	} else {
		// S_IFREG だとファイル
		stat.Mode = fuse.S_IFREG | 0444
		// 属性を設定
		stat.Size = a.Size
		stat.Mtim = fuse.NewTimespec(a.ModTime)
	}

	// ハードリンクされてる数 (当然1)
	stat.Nlink = 1

	return 0
}
