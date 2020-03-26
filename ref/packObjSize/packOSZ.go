package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/a4a881d4/gitcrawling/badgerdb"
	"github.com/a4a881d4/gitcrawling/gitext"
	"github.com/a4a881d4/gitcrawling/packext"
)

var (
	argDir  = flag.String("o", "../temp/.gitdb", "The dir story")
	argDump = flag.Bool("d", false, "dump Trees")
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

	if *argDump {
		var lastk, lastv []byte
		lastk = make([]byte, 45)
		tdb.ForEach([]byte("hash/"), func(k, v []byte) error {
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
	} else {
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
}
