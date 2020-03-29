package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/a4a881d4/gitcrawling/badgerdb"
	"github.com/a4a881d4/gitcrawling/gitext"
	"github.com/a4a881d4/gitcrawling/packext"
)

var (
	argDBDir   = flag.String("gitdb", "../temp/.gitdb", "The dir of db")
	argDestDir = flag.String("dest", "../temp/copy", "The dir of db")
	argMod     = flag.String("m", "import", "mode import,count,dump,dedup")
	argGrp     = flag.Int("g", 256, "split")
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
	switch *argMod {
	case "json":
		Json(tdb)
	case "import":
		importObj(tdb)
	default:
		importObj(tdb)
		Json256(tdb)
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

func Json(tdb *badgerdb.DB) {
	var format string
	var g int
	switch *argGrp {
	case 16:
		format = "%01x"
		g = 16
	case 256:
		format = "%02x"
		g = 256
	case 4096:
		format = "%03x"
		g = 4096
	default:
		format = "%02x"
		g = 256
	}
	for i := 0; i < g; i++ {
		pf := fmt.Sprintf(format, i)
		err := build(tdb, pf)
		if err != nil {
			fmt.Println(err)
		}
	}
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

	wrfn := fmt.Sprintf("%s/%s.json", *argDestDir, prefix)
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

func importObj(tdb *badgerdb.DB) {
	tdb.NewSession()
	defer tdb.EndSession()

	var doSome = func(fn string) {
		op, r, err := gitext.GetOffsetNoClassify(fn)
		tdb.Put([]byte("file/"+op.String()), []byte(fn))
		if err != nil {
			fmt.Println(err)
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
