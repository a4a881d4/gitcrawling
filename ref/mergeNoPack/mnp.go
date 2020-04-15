package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/a4a881d4/gitcrawling/packext"
)

func main() {
	err := packext.DefaultFromDir(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	nopack, err := packext.NewMergeNoFile(os.Args[2])
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, v := range packext.DefaultMap {
		idxf := strings.Replace(v, ".pack", ".idx", -1)
		fmt.Println("Doing", idxf)
		err = nopack.AddFile(idxf)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
