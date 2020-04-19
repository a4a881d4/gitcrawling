package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/a4a881d4/gitcrawling/packext"
)

func main() {
	dir := os.Args[1]
	stat, err := os.Stat(dir)
	if err != nil {
		fmt.Println(err)
		return
	}
	if stat.IsDir() {
		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			fn := filepath.Base(path)
			if strings.Contains(fn, ".pack") && strings.Contains(fn, "packn-") {
				err := packext.BuildIdxFile(path)
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			fmt.Println(err)
		}
	} else {
		fmt.Printf("%s must be dir", dir)
	}
}
