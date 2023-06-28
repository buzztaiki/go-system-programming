package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/edsrzf/mmap-go"
)

func loop(m mmap.MMap) {
	stdin := bufio.NewScanner(os.Stdin)
	for stdin.Scan() {
		cmdline := strings.Split(strings.TrimSpace(stdin.Text()), " ")

		cmd := cmdline[0]
		switch cmd {
		case "":
			return
		case "w":
			if len(cmdline) != 3 {
				fmt.Fprintf(os.Stderr, "invalid cmdline %s\n", strings.Join(cmdline, " "))
				continue
			}
			index, _ := strconv.Atoi(cmdline[1])
			ch := cmdline[2][0]
			m[index] = ch

			fmt.Printf("m[%d] = %q\n%s\n", index, ch, m)
		case "r":
			fmt.Printf("%s\n", m)
		default:
			fmt.Fprintf(os.Stderr, "invalid cmdline %s\n", strings.Join(cmdline, " "))
		}
	}
}

func open() (*os.File, error) {
	testPath := filepath.Join(os.TempDir(), "mmap_multi_testdata")
	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		fmt.Println("not exist")
		testData := []byte("0123456789")
		if err := ioutil.WriteFile(testPath, testData, 0644); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	f, err := os.OpenFile(testPath, os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	return f, nil
}

func prot(name string) int {
	switch name {
	case "rw":
		return mmap.RDWR
	case "r":
		return mmap.RDONLY
	case "cow":
		return mmap.COPY
	default:
		return mmap.RDWR

	}
}

func main() {
	protName := flag.String("prot", "rw", "(rw|r|cow)")
	flag.Parse()

	f, err := open()
	if err != nil {
		panic(err)
	}
	defer f.Close()

	m, err := mmap.Map(f, prot(*protName), 0)
	if err != nil {
		panic(err)
	}
	defer m.Unmap()

	loop(m)
	m.Flush()
}
