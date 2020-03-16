package main

import (
	"fmt"

	"github.com/a4a881d4/gitcrawling/badgerdb"
	"github.com/a4a881d4/gitcrawling/gitext"
)

func dumpCommit() {
	tdb, err := badgerdb.NewDB(*argDir + "/trees")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tdb.Close()

	tdb.ForEach([]byte("c"), func(k, v []byte) error {
		c, err := gitext.ObjToCommit(v)
		if err != nil {
			return err
		}
		copy(c.Hash[:], k[1:])
		fmt.Println(c)
		return nil
	})
}
