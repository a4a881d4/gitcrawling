package main

import (
	"flag"
	"fmt"
	"path"

	"github.com/a4a881d4/gitcrawling/packext"

	"github.com/a4a881d4/gitcrawling/badgerdb"
)

var (
	argDir = flag.String("o", "../temp", "The dir story")
	argMod = flag.String("m", "commit", "mode commit,tree,blob,blobsplit")
	all    = 0
)

func main() {
	flag.Parse()

	tdb, err := badgerdb.NewDB(path.Join(*argDir, ".gitdb", "objs"))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tdb.Close()
	packext.SortMode = "file"
	m, err := tdb.Group(3, 1)
	if err != nil {
		fmt.Println(err)
		return
	}
	WriteToPack(m, *argMod, tdb)
}

func WriteToPack(m map[string][]string, t string, tdb *badgerdb.DB) {
	pf, err := packext.NewFileDirPFDB(tdb, path.Join(*argDir, "packs"))
	if err != nil {
		fmt.Println(err)
		return
	}

	g := packext.NewObjectGet(pf)
	if objs, ok := m[t]; ok {
		if s, err := packext.NewSelectFile(path.Join(*argDir, t), objs, g); err != nil {
			fmt.Println(err)
			return
		} else {
			if err = packext.Flush(s); err != nil {
				fmt.Println(err)
				return
			}
		}
	} else {
		fmt.Println("unsupport", t)
	}
}
