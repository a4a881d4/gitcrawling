package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/a4a881d4/gitcrawling/db"
	"github.com/a4a881d4/gitcrawling/gitext"
)

var (
	argReposDir = flag.String("r", ".gitdb", "The dir story Repos")
	argRefsDir  = flag.String("ref", ".gitdb", "The dir story Refs")
	argForce    = flag.Bool("f", false, "force re clone")
	argGithub   = flag.String("g", "github.com", "github server")
	argThread   = flag.Int("t", 0, "Multi thread clone")
)
var (
	token chan int
	done  int
	wg    sync.WaitGroup
)

func main() {
	flag.Parse()

	token = make(chan int, *argThread)

	rdb := db.NewRefDB(*argRefsDir + "/refs")

	var doSome = func(names []string) {
		rdb.CashePrefech(names)
		rdb.NoDB()
		for num, name := range names {
			repo := strings.Split(name, "/")
			if len(repo) != 2 {
				fmt.Println("error name", name)
				continue
			}
			owner, project := repo[0], repo[1]
			if rdb.OK(owner, project) {
				continue
			}
			fmt.Println("Begin to Clone", owner, project, num, done, time.Now())
			done++
			url := fmt.Sprintf("http://%s/%s/%s", *argGithub, owner, project)
			path := fmt.Sprintf("%s/repos/%s/%s", *argReposDir, owner, project)
			_, err := os.Stat(path)
			if err == nil {
				if *argForce {
					os.RemoveAll(path)
				} else {
					continue
				}
			}
			r, err := gitext.PlainCloneFS(url, path)
			if err != nil {
				fmt.Println(err)
				continue
			}

			ref, err := r.Head()
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("HEAD: ", ref.Hash().String())
			refs := gitext.RepoRef(r)
			rdb.PutRef(owner, project, refs)
			dump(refs)
			fmt.Println("End ", owner, project, num, done, time.Now())
		}
		rdb.UnCashe(names)
		<-token
		fmt.Println("Done")
		wg.Done()
	}
	batchDo(doSome)
	fmt.Println("Wait Clone finish")
	wg.Wait()
	rdb.Stop()
}

func dump(ref []gitext.Ref) {
	if len(ref) > 3 {
		ref = ref[:3]
	}
	for k, v := range ref {
		fmt.Println(k, ":", v)
	}
}

func batchDo(putSome func([]string)) {
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
	var batch []string
	for _, v := range rec {
		for _, name := range v {
			batch = append(batch, name)
			if len(batch) > 32 {
				go putSome(batch)
				wg.Add(1)
				token <- 1
				batch = []string{}
			}
		}
	}
	if len(batch) > 0 {
		go putSome(batch)
		wg.Add(1)
		token <- 1
	}
}
