package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/a4a881d4/gitcrawling/badgerdb"
	"github.com/a4a881d4/gitcrawling/gitext"
)

func Raw() {
	var err error

	if *argDump {
		tdb, err := badgerdb.NewDB(*argDir + "/trees")
		if err != nil {
			fmt.Println(err)
			return
		}
		tdb.RawTrees(func(k, v []byte) error {
			fmt.Println(hex.EncodeToString(k[1:]))
			entries, err := gitext.DumpTree(v)
			if err != nil {
				fmt.Println(err)
			}
			for _, e := range entries {
				fmt.Println(gitext.TreeEntry2String(e))
			}
			return nil
		})
		tdb.Close()
		rdb, err := badgerdb.NewDB(*argDir + "/refs")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer rdb.Close()
		rdb.RawRefs(func(k, v []byte) error {
			fmt.Println(string(k))
			ReadDB(hex.EncodeToString(v))
			return nil
		})
		return
	}
	tdb, err := badgerdb.NewDB(*argDir + "/trees")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tdb.Close()

	rdb, err := badgerdb.NewDB(*argDir + "/refs")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rdb.Close()

	buf, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Println(err)
		return
	}

	var rec = make(map[string][]string)
	err = json.Unmarshal(buf, &rec)
	if err != nil {
		fmt.Println(err)
		return
	}

	var putSome = func(names []string) {
		tdb.NewSession()
		defer tdb.EndSession()
		rdb.NewSession()
		defer rdb.EndSession()
		for _, name := range names {
			repo := strings.Split(name, "/")
			if len(repo) != 2 {
				fmt.Println("error name", name)
				continue
			}
			owner, project := repo[0], repo[1]
			path := fmt.Sprintf("%s/repos/%s/%s", *argReposDir, owner, project)
			_, err := os.Stat(path)
			if err != nil {
				fmt.Println(ShowName(owner, project), "miss")
				missNum++
			} else {
				path := fmt.Sprintf("%s/repos/%s/%s", *argReposDir, owner, project)
				fmt.Println(ShowName(owner, project), "check", path)
				r, err := gitext.PlainOpen(path)
				if err != nil {
					fmt.Println(ShowName(owner, project), err)
				} else {
					err = gitext.Trees(r, tdb.PutRawTree, rdb.PutRawRef)
					if err != nil {
						fmt.Println(err)
					}
				}
			}
			repoNum++
		}
	}
	var batch []string
	for _, v := range rec {
		for _, name := range v {
			batch = append(batch, name)
		}
		if len(batch) > 2048 {
			putSome(batch)
			batch = []string{}
		}
	}
	if len(batch) > 0 {
		putSome(batch)
	}
	fmt.Printf("%8d/%d\n", repoNum-missNum, repoNum)
}
