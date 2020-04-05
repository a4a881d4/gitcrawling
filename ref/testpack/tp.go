package main

import (
	"github.com/a4a881d4/gitcrawling/types"
	"encoding/hex"
	"github.com/a4a881d4/gitcrawling/packext"
	"flag"
	"fmt"

	"github.com/a4a881d4/gitcrawling/badgerdb"

)

var (
	argDir  = flag.String("o", "../temp/.gitdb", "The dir story")
	argMod  = flag.String("m", "ls", "mode import,count,dump,dedup")
	all     = 0
	dup     = 0
)

func main() {
	flag.Parse()

	tdb, err := badgerdb.NewDB(*argDir + "/.gitdb/objs")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tdb.Close()

	switch *argMod {

	case "ls":
		ls(tdb)
	default:
		Hash(tdb)
	}
}

func Hash(tdb *badgerdb.DB) {

	pf,err := packext.NewFileDirPFDB(tdb,*argDir)
	if err != nil {
		fmt.Println(err)
		return
	}
	h,err := hex.DecodeString(flag.Arg(0))
	if err != nil {
		fmt.Println(err)
		return
	}
	var hash types.Hash
	copy(hash[:],h[:])
	r,err := pf.Get(hash)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(len(r))
	s := packext.NewScanner(r)
	head,err := s.Header()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(head)
}

func ls(tdb *badgerdb.DB) {
	prefix := []byte(flag.Arg(0))
	tdb.ForEach(prefix, func(k, v []byte) error {
		fmt.Println(string(k), ":", string(v))
		return nil
	})
}