package main

import (
	"context"
	"io"
	"log"
	"strings"

	"github.com/billziss-gh/cgofuse/fuse"
)

func (cf *CloudFileSystem) Read(path string, buff []byte, ofst int64, fh uint64) int {
	name := strings.TrimLeft(path, "/")
	ctx := context.Background()
	reader, err := cf.bucket.NewRangeReader(ctx, name, ofst, int64(len(buff)), nil)
	if err != nil {
		return 0
	}
	defer reader.Close()

	n, err := reader.Read(buff)
	if err != nil && err != io.EOF {
		log.Printf("failed to erad path: %v", err)
		return fuse.EIO
	}
	return n
}
