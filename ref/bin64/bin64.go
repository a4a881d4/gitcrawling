package main

import (
	"encoding/base64"
	"fmt"
	"os"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

func main() {
	r, _ := git.PlainOpen(os.Args[1])

	root, _ := r.Head()

	hash := root.Hash()

	b64 := base64.RawURLEncoding.EncodeToString(hash[:])

	rehash, err := base64.RawURLEncoding.DecodeString(b64)

	if err != nil {
		fmt.Println(err)
	}

	var rh plumbing.Hash

	copy(rh[:], rehash[:])

	fmt.Println(b64, rh.String(), hash.String())

}
