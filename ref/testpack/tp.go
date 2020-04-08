package main

import (
	"flag"
	"fmt"
	"path"

	"github.com/a4a881d4/gitcrawling/packext"
	"github.com/a4a881d4/gitcrawling/types"

	"github.com/a4a881d4/gitcrawling/badgerdb"
)

var (
	argDir = flag.String("o", "../temp", "The dir story")
	argMod = flag.String("m", "ls", "mode import,count,dump,dedup")
	all    = 0
	dup    = 0
)

func main() {
	flag.Parse()

	tdb, err := badgerdb.NewDB(path.Join(*argDir, ".gitdb", "objs"))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tdb.Close()

	switch *argMod {

	case "ls":
		ls(tdb, []byte(flag.Arg(0)))
	default:
		Hash(tdb)
	}
}

func Hash(tdb *badgerdb.DB) {
	oes := ls(tdb, []byte("hash/"+flag.Arg(0)))

	pf, err := packext.NewFileDirPFDB(tdb, *argDir)
	if err != nil {
		fmt.Println(err)
		return
	}

	s := packext.NewObjectGet(pf)
	for _, oe := range oes {
		head, err := s.HeaderByHash(types.Hash(oe.Hash))
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(head)
	}

}

func ls(tdb *badgerdb.DB, prefix []byte) []*packext.ObjEntry {
	var res = []*packext.ObjEntry{}
	tdb.ForEach(prefix, func(k, v []byte) error {
		fmt.Println(string(k), ":", string(v))
		var oe = packext.ObjEntry{}
		oe.SetKey(k)
		res = append(res, &oe)
		return nil
	})
	return res
}
