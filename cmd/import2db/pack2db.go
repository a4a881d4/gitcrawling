package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/a4a881d4/gitcrawling/badgerdb"
	"github.com/a4a881d4/gitcrawling/gitext"
	"github.com/a4a881d4/gitcrawling/objext"
	"github.com/a4a881d4/gitcrawling/packext"
	"github.com/a4a881d4/gitcrawling/types"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

var (
	argDir = flag.String("o", "../temp/.gitdb", "The dir story")
	argMod = flag.String("m", "import", "mode import,ls")
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

	switch *argMod {
	case "ls":
		ls(tdb, []byte(flag.Arg(0)))
	case "import":
		importObj(tdb)
	case "cat":
		catObj(tdb)
	case "hash":
		Hash(tdb)
	case "hashfile":
		HashFile(tdb)

	default:
		ls(tdb, []byte(flag.Arg(0)))
	}
}

func importObj(tdb *badgerdb.DB) {
	tdb.NewSession()
	defer tdb.EndSession()

	var doSome = func(fn string) {
		op, r, err := gitext.GetOffsetNoClassify(fn)
		tdb.Put([]byte("file/"+op.String()), []byte(fn))
		if err != nil {
			fmt.Println(err, fn)
		}
		for _, e := range r {
			err = tdb.BPut(&e)
			if err != nil {
				fmt.Println(err)
			}
			all += 1
		}
	}
	stat, err := os.Stat(flag.Arg(0))
	if err != nil {
		fmt.Println(err)
		return
	}
	if stat.IsDir() {
		filepath.Walk(flag.Arg(0), func(path string, info os.FileInfo, err error) error {
			if strings.Contains(path, ".idx") && strings.Contains(path, "pack-") {
				doSome(path)
			}
			return nil
		})
	} else {
		doSome(flag.Arg(0))
	}
	fmt.Println("All:", all)
}

func ls(tdb *badgerdb.DB, prefix []byte) []*packext.ObjEntry {
	var res = []*packext.ObjEntry{}
	tdb.ForEach(prefix, func(k, v []byte) error {
		var oe = packext.ObjEntry{}
		oe.SetKey(k)
		res = append(res, &oe)
		return nil
	})
	return res
}

func catObj(tdb *badgerdb.DB) {
	prefix := "hash/"
	if len(flag.Args()) != 0 {
		prefix += flag.Arg(0)
	}
	oes := ls(tdb, []byte(prefix))

	pf, err := packext.NewFileDirPFDB(tdb, *argDir)
	if err != nil {
		fmt.Println(err)
		return
	}

	s := packext.NewObjectGet(pf)
	for _, oe := range oes {
		body, err := s.Body(types.Hash(oe.Hash))
		if err != nil {
			fmt.Println(err)
			return
		}
		o := objext.NewBytesObject(body, oe)
		switch oe.RealType {
		case plumbing.CommitObject:
			c, err := objext.DecodeCommit(o)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(c)
			}
		case plumbing.TreeObject:
			t, err := objext.DecodeTree(o)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(oe.RealType.String(), oe.Hash.String())
				ts := objext.Tree2String(t)
				fmt.Println(ts)
			}
		}
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
		fmt.Println(oe, oe.OHeader)
		fmt.Println(head)
	}
}

func HashFile(tdb *badgerdb.DB) {
	var m, mbase map[string][]string
	var err error
	if m, err = tdb.Group(3, 1); err != nil {
		fmt.Println(err)
	} else {
		var js []byte
		if js, err = json.MarshalIndent(m, "", "  "); err != nil {
			fmt.Println(err)
		} else {
			if err = ioutil.WriteFile(path.Join(*argDir, ".gitdb", "hash"), js, 0644); err != nil {
				fmt.Println(err)
			}
		}
	}
	if mbase, err = tdb.Group(3, 4); err != nil {
		fmt.Println(err)
	} else {
		var js []byte
		if js, err = json.MarshalIndent(mbase, "", "  "); err != nil {
			fmt.Println(err)
		} else {
			if err = ioutil.WriteFile(path.Join(*argDir, ".gitdb", "base"), js, 0644); err != nil {
				fmt.Println(err)
			}
		}
		for k, v := range mbase {
			dedup := make(map[string]bool)
			for _, ks := range m[k] {
				dedup[ks] = true
			}
			for _, bs := range v {
				if _, ok := dedup[bs]; !ok {
					fmt.Println("Miss delta base", bs)
				}
			}
		}
	}
}
