package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/a4a881d4/gitcrawling/gitext"
)

func main() {
	dir, err := ioutil.ReadDir(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	m, err := gitext.NewMergeFile(os.Args[2])
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, v := range dir {
		if !v.IsDir() && strings.Contains(v.Name(), ".pack") {
			inf := path.Join(os.Args[1], v.Name())
			err = m.AddFile(inf, uint64(1<<28))
			if err == nil {
				continue
			}
			if err == gitext.ErrOutSize {
				err = m.Flush()
				if err != nil {
					fmt.Println(err)
					return
				}
				m, err = gitext.NewMergeFile(os.Args[2])
				if err != nil {
					fmt.Println(err)
					return
				}
			} else {
				fmt.Println(err)
				return
			}
		}
	}
	err = m.Flush()
	if err != nil {
		fmt.Println(err)
		return
	}
}
