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
	// packext.SortMode = "file"
	m, err := tdb.Group(3, 1)
	if err != nil {
		fmt.Println(err)
		return
	}
	WriteToPack(m, *argMod, tdb)
}
func dedup(os []string, filemap map[string][]string) []string {
	var r = []string{}
	var m = make(map[string]bool)
	for _, o := range os {
		m[o] = true
	}
	for _, ofs := range filemap {
		for _, o := range ofs {
			if _, ok := m[o]; ok && m[o] {
				r = append(r, o)
				m[o] = false
			}
		}
	}

	for k, v := range m {
		if v {
			fmt.Println("Miss", k)
		}
	}
	return r
}

func WriteToPack(m map[string][]string, t string, tdb *badgerdb.DB) {
	pf, err := packext.NewFileDirPFDB(tdb, path.Join(*argDir, "packs"))
	if err != nil {
		fmt.Println(err)
		return
	}
	filemap, err := tdb.Group(2, 1)
	if err != nil {
		fmt.Println(err)
		return
	}
	g := packext.NewObjectGet(pf)
	if objs, ok := m[t]; ok {
		objs = dedup(objs, filemap)
		s, err := packext.NewSelectFile(path.Join(*argDir, t), objs, g)
		if err != nil {
			fmt.Println(err)
			return
		}
		if err = packext.Flush(s); err != nil {
			fmt.Println(err)
			return
		}
	} else {
		fmt.Println("unsupport", t)
	}
}
