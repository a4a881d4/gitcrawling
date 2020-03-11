package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/a4a881d4/gitcrawling/db"
	"github.com/a4a881d4/gitcrawling/gitext"
)

var (
	argReposDir = flag.String("r",".gitdb","The dir story Repos")
	argRefsDir  = flag.String("ref",".gitdb","The dir story Refs")
	argSleep    = flag.Int("s",1000,"sleep")
	argCompact  = flag.Bool("c",false,"Compact DB before dump")
	repoNum     = 0
	missNum     = 0
)

func main() {
	flag.Parse()
	var err error

	if *argCompact {
		err := db.Compact(*argRefsDir+"/refs")
		if err != nil {
			fmt.Println(err)
		}
	}

	rdb := db.NewRefDB(*argRefsDir+"/refs")
	
	buf,err := ioutil.ReadFile(flag.Arg(0))
	if err!=nil {
		fmt.Println(err)
		return
	}

	var rec = make(map[string][]string)
	err = json.Unmarshal(buf,&rec)
	if err!=nil {
		fmt.Println(err)
		return
	}
	for _,v := range rec {
		for _,name := range v {
			repo := strings.Split(name,"/")
			if len(repo)!=2 {
				fmt.Println("error name",name)
				continue
			}
			owner,project := repo[0],repo[1]
			path := fmt.Sprintf("%s/repos/%s/%s",*argReposDir,owner,project)
			_, err := os.Stat(path)
			if err!=nil {
				fmt.Println(ShowName(owner,project),"miss")
				missNum++
				if rdb.OK(owner,project) {
					fmt.Println(ShowName(owner,project),"bad in db, remove it")
					rdb.DelRef(owner,project)
				}
			} else {
				if rdb.IsBuild(owner,project) {
					fmt.Println(ShowName(owner,project),"has build")
				}
				if rdb.OK(owner,project) {
					fmt.Println(ShowName(owner,project),"clone local")
				} else {
					fmt.Println(ShowName(owner,project),"maybe clone local, check")
					refs,err := OpenAndRef(owner,project,*argReposDir)
					if err != nil || len(refs) == 0 {
						fmt.Println(ShowName(owner,project),"bad",path,"remove it")
						os.RemoveAll(path)
						rdb.DelRef(owner,project)
					} else {
						rdb.PutRefSync(owner,project,refs)
						dump(refs)
					}
				}
			}
			<- time.After(time.Duration(*argSleep) * time.Millisecond)
			repoNum++
		}
	}
	fmt.Printf("%8d/%d\n",repoNum-missNum,repoNum)
}

func ShowName(owner,project string) string {
	var space = "                                                                        "
	num := fmt.Sprintf("%8d ",repoNum)
	if len(owner) > 25 {
		owner = owner[:25]
	} else {
		owner = space[:25-len(owner)]+owner
	}
	if len(project) > 35 {
		project = project[:35]
	} else {
		project = project + space[:35-len(project)]
	}
	return num+owner+":"+project
}

func OpenAndRef(owner,project,ReposDir string) (ref []gitext.Ref,err error) {
	path  := fmt.Sprintf("%s/repos/%s/%s",ReposDir,owner,project)
	r,err := gitext.PlainOpen(path)
	if err != nil {
		return []gitext.Ref{},err
	}
	return gitext.RepoRef(r),err
}

func dump(ref []gitext.Ref) {
	for k,v := range ref {
		fmt.Println(k,":",v)
	}
}