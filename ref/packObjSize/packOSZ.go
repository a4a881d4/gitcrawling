package main

import (
	"fmt"
	"os"

	"github.com/a4a881d4/gitcrawling/gitext"
)

func main() {
	r, err := gitext.GetOffset(os.Args[1])
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(r)
}
