package main

import (
	"os"
	"fmt"
	"github.com/a4a881d4/gitcrawling/gitext"
)



func main() {
	packf := os.Args[1]
	hs,err := gitext.Reidx(packf)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(hs)
	}
}