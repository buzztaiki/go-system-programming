package main

import (
	"sync"
	"syscall"

	"golang.org/x/sys/windows"
)

type FileLock struct {
	m  sync.Mutex
	fd windows.Handle
}

func NewFileLock(filename string) *FileLock {
	if filename == "" {
		panic("filename needed")
	}
	fd, err := windows.CreateFile(
		windows.StringToUTF16Ptr(filename),
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		windows.OPEN_ALWAYS,
		windows.FILE_ATTRIBUTE_NORMAL,
		0)
	if err != nil {
		panic(err)
	}
	return &FileLock{fd: fd}
}

func (m *FileLock) Lock() {
	m.m.Lock()
	err := windows.LockFileEx(
		m.fd,
		windows.LOCKFILE_EXCLUSIVE_LOCK,
		0,
		1,
		0,
		&windows.Overlapped{})
	if err != nil {
		panic(err)
	}
}

func (m *FileLock) Unlock() {
	err := windows.UnlockFileEx(
		m.fd,
		0,
		1,
		0,
		&windows.Overlapped{})
	if err != nil {
		panic(err)
	}
	m.m.Unlock()
}
