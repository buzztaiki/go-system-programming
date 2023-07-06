package main

import (
	"context"
	"io"
	"os"

	"gocloud.dev/blob"

	// init 関数だけ呼んで実際の関数は使わない場合こういう書き方をする
	// 中でサービスプロバイダーみたいな人に登録してるケースが多い
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"
)

func main() {
	bucketUrl := os.Args[1]
	path := os.Args[2]

	ctx := context.Background()
	bucket, err := blob.OpenBucket(ctx, bucketUrl)
	defer bucket.Close()
	if err != nil {
		panic(err)
	}

	reader, err := bucket.NewReader(ctx, path, nil)
	defer reader.Close()
	if err != nil {
		panic(err)
	}

	io.Copy(os.Stdout, reader)
}
