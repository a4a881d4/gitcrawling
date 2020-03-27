package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/a4a881d4/gitcrawling/badgerdb"
	"github.com/a4a881d4/gitcrawling/gitext"
	"github.com/a4a881d4/gitcrawling/packext"
)

var (
	argDir  = flag.String("o", "../temp/.gitdb", "The dir story")
	argMod  = flag.String("m", "import", "mode import,count,dump,dedup")
	argTemp = flag.String("t", "../temp", "temp dir")
	all     = 0
	dup     = 0
)

func main() {
	flag.Parse()
	tdb, err := badgerdb.NewDB(*argDir + "/objs")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tdb.Close()

	switch *argMod {
	case "dump":
		dumpObj(tdb)
	case "import":
		importObj(tdb)
	case "dedup":
		deDupObj(tdb)
	case "count":
		countObj(tdb)
	default:
		dumpObj(tdb)
	}
}
func countObj(tdb *badgerdb.DB) {
	var counter = make([]int64, 33)
	var total int64 = 0
	var count = func(a uint32) int {
		var r int = 0
		for ; a != 0; r++ {
			a = a >> 1
		}
		counter[r]++
		total++
		return r
	}
	tdb.ForEach([]byte("hash/"), func(k, v []byte) error {
		b := &packext.ObjEntry{}
		b.FromByte(v)
		count(b.Size)
		return nil
	})

	for i, v := range counter {
		if i == 0 {
			fmt.Printf("%10d: %8d ", 0, v*1_000_000/total)
		} else {
			fmt.Printf("%10d: %8d ", 1<<(i-1), v*1_000_000/total)
		}
		if i&3 == 3 {
			fmt.Println()
		}
	}
}

func deDupObj(tdb *badgerdb.DB) {

	var lastk []byte
	lastk = make([]byte, 45)
	b := &packext.ObjEntry{}
	a := &packext.ObjEntry{}
	tmpfile := path.Join(*argTemp, "tmp-dup-key")
	dupf, err := os.Create(tmpfile)
	if err != nil {
		fmt.Println(err)
		return
	}
	tdb.ForEachOne([]byte("hash/"), func(k, v []byte) error {
		if bytes.Equal(lastk[:45], k[:45]) {
			a.FromByte(v)
			if a.Size < b.Size {
				lastk, k = k, lastk
				a, b = b, a
			}
			_, err := fmt.Fprintln(dupf, string(k))
			if err != nil {
				fmt.Println(err)
			}
			dup += 1
		} else {
			lastk, k = k, lastk
			a, b = b, a
		}
		return nil
	})
	dupf.Close()
	fmt.Println("Dup:", dup)

	tdb.NewSession()
	defer tdb.EndSession()

	dupf, err = os.Open(tmpfile)
	if err != nil {
		fmt.Println(err)
		return
	}
	buf := bufio.NewReader(dupf)
	for {
		line, err := buf.ReadString('\n')
		if err == nil || err == io.EOF {
			line = strings.TrimSpace(line)
			erri := tdb.Delete([]byte(line))
			if erri != nil {
				fmt.Println(erri)
			}
			if err == io.EOF {
				return
			}
		} else {
			fmt.Println(err)
			return
		}
	}
}
func importObj(tdb *badgerdb.DB) {
	tdb.NewSession()
	defer tdb.EndSession()

	var doSome = func(fn string) {
		r, err := gitext.GetOffsetNoClassify(fn)
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
			if strings.Contains(path, ".idx") {
				doSome(path)
			}
			return nil
		})
	} else {
		doSome(flag.Arg(0))
	}
	fmt.Println("All:", all)
}
func dumpObj(tdb *badgerdb.DB) {
	var lastk, lastv []byte
	lastk = make([]byte, 45)
	tdb.ForEachOne([]byte("hash/"), func(k, v []byte) error {
		if bytes.Equal(lastk[:45], k[:45]) {
			a := &packext.ObjEntry{}
			b := &packext.ObjEntry{}
			a.FromByte(lastv)
			b.FromByte(v)
			fmt.Println(string(lastk), a.CRC32, a.Size)
			fmt.Println(string(k), b.CRC32, b.Size)
			fmt.Println()
			dup += 1
		}
		lastk = k
		lastv = v
		return nil
	})
	fmt.Println("Dup:", dup)
}
