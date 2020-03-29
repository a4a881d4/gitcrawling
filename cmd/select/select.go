package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/a4a881d4/gitcrawling/badgerdb"
	"github.com/a4a881d4/gitcrawling/packext"
)

var (
	argDBDir   = flag.String("gitdb", "../temp/.gitdb", "The dir of db")
	argDestDir = flag.String("dest", "../temp/copy", "The dir of db")
	all        = 0
	dup        = 0
)

func main() {
	flag.Parse()
	tdb, err := badgerdb.NewDB(*argDBDir + "/objs")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tdb.Close()
	for i := 0; i < 256; i++ {
		pf := fmt.Sprintf("%02x", i)
		err = build(tdb, pf)
		if err != nil {
			fmt.Println(err)
		}
	}
}

type OE struct {
	O uint64
	S uint32
	C uint32
}

type packfile struct {
	fd       io.ReaderAt
	FileName string
	Task     []OE
}

func getFileFromDB(tdb *badgerdb.DB) (map[packext.OriginPackFile]*packfile, error) {
	files := make(map[packext.OriginPackFile]*packfile)
	err := tdb.ForEach([]byte("file/"), func(k, v []byte) error {
		var h packext.OriginPackFile
		hs, err := hex.DecodeString(string(k[5:]))
		if err != nil {
			return err
		}
		copy(h[:], hs[:])
		files[h] = &packfile{
			FileName: string(v),
		}
		return nil
	})
	return files, err
}

func build(tdb *badgerdb.DB, prefix string) error {
	files, err := getFileFromDB(tdb)
	if err != nil {
		return err
	}
	s := tdb.NewHashSession(prefix)
	defer s.End()
	var newEntry = func() badgerdb.Byter {
		return &packext.ObjEntry{}
	}
	var small *packext.ObjEntry
	for {
		items, err := s.NextGroup(45, newEntry)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if len(items) > 0 {
			for k, v := range items {
				o := v.(*packext.ObjEntry)
				if k == 0 {
					small = o
				} else {
					if o.Size < small.Size {
						small = o
					}
				}
			}
			dup++
			all += len(items)
			if _, ok := files[small.PackFile]; ok {
				files[small.PackFile].Task = append(files[small.PackFile].Task,
					OE{small.Offset, small.Size, small.CRC32})
			}
		}
	}
	fmt.Println("Res:", all, dup)

	wrfn := fmt.Sprintf("%s/%s.json", *argDestDir, flag.Arg(0))
	w, err := os.Create(wrfn)
	if err != nil {
		return err
	}
	defer w.Close()
	enc := json.NewEncoder(w)
	if err != nil {
		return err
	}
	enc.SetIndent("", "  ")
	for _, v := range files {
		err = enc.Encode(v)
		if err != nil {
			return err
		}
	}
	return nil
}
