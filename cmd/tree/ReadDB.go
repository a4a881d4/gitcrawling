package main

import (
	"encoding/hex"
	"fmt"

	"github.com/a4a881d4/gitcrawling/badgerdb"
	"github.com/a4a881d4/gitcrawling/gitext"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func ReadDB(hash string) {
	h, err := hex.DecodeString(hash)
	if err != nil {
		fmt.Println(err)
		return
	}
	var H plumbing.Hash
	copy(H[:], h)
	tdb, err := badgerdb.NewDB(*argDir + "/trees")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tdb.Close()
	m, err := gitext.Tree(H, tdb)
	if err != nil {
		fmt.Println(err)
		return
	}
	for k, v := range m {
		fmt.Println(gitext.TreeEntry2String(&object.TreeEntry{
			Hash: k,
			Name: v.Name,
			Mode: v.Mode,
		}))
	}
}
