package main

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/src-d/go-git.v4/plumbing/format/idxfile"
)

func main() {

	r, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println(1, err)
	}
	defer r.Close()

	dec := idxfile.NewDecoder(r)
	idx := idxfile.NewMemoryIndex()

	err = dec.Decode(idx)
	if err != nil {
		fmt.Println(2, err)
	}

	iter, err := idx.Entries()
	if err != nil {
		fmt.Println(3, err)
	}

	for {
		e, err := iter.Next()
		if err == io.EOF {
			break
		}

		fmt.Println(e.Hash.String(), e.Offset)
	}
}
