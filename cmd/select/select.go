package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/a4a881d4/gitcrawling/badgerdb"
	"github.com/a4a881d4/gitcrawling/gitext"
	"github.com/a4a881d4/gitcrawling/packext"
	"gopkg.in/src-d/go-git.v4/plumbing"
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
		toJson(tdb)
	case "import":
		importObj(tdb)
	case "build":
		fromJson()
	default:
		importObj(tdb)
		toJson(tdb)
		fromJson()
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

func toJson(tdb *badgerdb.DB) {
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
		err := buildToJson(tdb, pf)
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

func buildToJson(tdb *badgerdb.DB, prefix string) error {
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

func fromJson() {
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
		pf = pf + ".json"
		jf := path.Join(*argDestDir, pf)
		err := buildFromJson(jf)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func buildFromJson(jf string) error {
	r, err := os.Open(jf)
	if err != nil {
		return err
	}
	defer r.Close()
	var packfiles []*packfile
	dec := json.NewDecoder(r)
	for {
		var p packfile
		err = dec.Decode(&p)
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			return err
		}
		packfiles = append(packfiles, &p)
	}
	objNum := 0
	for _, v := range packfiles {
		objNum += len(v.Task)
	}
	fmt.Println(jf, objNum)
	tempfn := path.Join(*argDestDir, "tmp-pack")
	var writeToTemp = func(tfn string) (h plumbing.Hash, terr error) {
		var pf *gitext.PackFile
		pf, terr = gitext.NewPack(tempfn)
		if terr != nil {
			return
		}
		defer pf.Close()

		terr = pf.Head(objNum)
		if terr != nil {
			return
		}
		var doOne = func(v *packfile) error {
			pfn := strings.Replace(v.FileName, ".idx", ".pack", -1)
			// fmt.Println("open", pfn)
			r, derr := os.Open(pfn)
			if derr != nil {
				return derr
			}
			defer r.Close()
			for _, oe := range v.Task {
				_, derr = r.Seek(int64(oe.O), 0)
				if derr != nil {
					fmt.Println(1)
					return derr
				}
				_, derr = pf.Do(r, int64(oe.S))
				if derr != nil {
					stat, _ := os.Stat(v.FileName)
					fmt.Println(2, oe.O, oe.S, stat.Size())
					return derr
				}
			}
			return nil
		}
		for _, v := range packfiles {
			terr = doOne(v)
			if terr != nil {
				fmt.Println(terr)
				continue
			}
		}
		h, terr = pf.Footer()
		if terr != nil {
			return
		}
		return
	}
	var hash plumbing.Hash
	hash, err = writeToTemp(tempfn)
	if err != nil {
		return err
	}
	newfn := path.Join(*argDestDir, "pack-"+hash.String()+".pack")
	err = os.Rename(tempfn, newfn)
	if err != nil {
		return err
	}
	cmd := exec.Command("git", "index-pack", newfn)
	err = cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
