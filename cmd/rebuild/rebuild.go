package main

import (
	"flag"
	"fmt"
	"path"
	"strings"

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
	// m, err := tdb.Group(3, 1)
	// if err != nil {
	// 	fmt.Println(0, err)
	// 	return
	// }
	WriteToPack(*argMod, tdb)
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

func WriteToPack(t string, tdb *badgerdb.DB) {
	pf, err := packext.NewFileDirPFDB(tdb, path.Join(*argDir, "packs"))
	if err != nil {
		fmt.Println(err)
		return
	}
	g := packext.NewObjectGet(pf)
	objs, err := Group(tdb, t)
	if err != nil {
		fmt.Println(err)
		return
	}
	s, err := packext.NewSelectFile(path.Join(*argDir, t), objs, g)
	if err != nil {
		fmt.Println(err)
		return
	}
	if err = packext.Flush(s); err != nil {
		fmt.Println(err)
		return
	}

}

func Group(tdb *badgerdb.DB, t string) ([]string, error) {

	r := make(map[string][]string)
	fmt.Println("Begin Group")
	fmt.Printf("Progress:\033[s")
	counter := 0
	err := tdb.ForEach([]byte("hash/"), func(k, v []byte) error {
		ss := strings.Split(string(k), "/")
		if ss[3] == t {
			r[ss[2]] = append(r[ss[2]], ss[1])
		}
		counter++
		if counter%100000 == 0 {
			fmt.Printf("\033[u\033[K%20d", counter)
		}
		return nil
	})
	fmt.Println("\nEnd Group")
	fmt.Println("Begin order")
	counter = 0
	fmt.Printf("Progress:\033[s")
	var os []string
	mo := make(map[string]bool)
	for _, v := range r {
		for _, o := range v {
			if _, ok := mo[o]; !ok {
				mo[o] = true
				os = append(os, o)
			}
			counter++
			if counter%100000 == 0 {
				fmt.Printf("\033[u\033[K%20d", counter)
			}
		}
	}
	fmt.Println("\nEnd order")
	return os, err
}
