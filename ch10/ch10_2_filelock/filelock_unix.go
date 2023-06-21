//go:build unix

package main

import (
	"os"
	"syscall"
)

func lock(f *os.File, op int) (err error) {
	for {
		err = syscall.Flock(int(f.Fd()), int(op))
		// https://blog.lufia.org/entry/2020/02/29/162727
		if err != syscall.EINTR {
			break
		}
	}
	return err
}

type FileLock struct {
	f *os.File
}

func NewFileLock(filename string) *FileLock {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	return &FileLock{f}
}

func (m *FileLock) Lock() {
	if err := lock(m.f, syscall.LOCK_EX); err != nil {
		panic(err)
	}
}

func (m *FileLock) Unlock() {
	if err := lock(m.f, syscall.LOCK_UN); err != nil {
		panic(err)
	}

}
