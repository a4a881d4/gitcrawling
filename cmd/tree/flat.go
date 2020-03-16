package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/a4a881d4/gitcrawling/badgerdb"
	"github.com/a4a881d4/gitcrawling/gitext"
)

func Flat() {
	var err error

	tdb, err := badgerdb.NewDB(*argDir + "/trees")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tdb.Close()
	if *argDump {
		prefix := "t/"
		if flag.NArg() == 1 {
			prefix += flag.Arg(0)
		}
		tdb.ForEach([]byte(prefix), func(k, v []byte) error {
			fmt.Println(string(k[2:]))
			fmt.Println(string(v))
			return nil
		})
		return
	}

	var putSome = func(names []string) {
		tdb.NewSession()
		defer tdb.EndSession()
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
				fmt.Println(ShowName(owner, project), "check")
				trees, err := OpenAndTree(owner, project, *argReposDir)
				if err != nil {
					fmt.Println(ShowName(owner, project), err)
				} else {
					tdb.PutTree(owner, project, trees)
				}
			}
			repoNum++
		}
	}
	batchDo(putSome)
	fmt.Printf("%8d/%d\n", repoNum-missNum, repoNum)
}

func OpenAndTree(owner, project, ReposDir string) (ref []string, err error) {
	path := fmt.Sprintf("%s/repos/%s/%s", ReposDir, owner, project)
	r, err := gitext.PlainOpen(path)
	if err != nil {
		return []string{}, err
	}
	return gitext.TreeFlat(r)
}
