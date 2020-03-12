package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/a4a881d4/gitcrawling/db"
	"github.com/a4a881d4/gitcrawling/gitext"
)

var (
	argReposDir = flag.String("r",".gitdb","The dir story Repos")
	argRefsDir  = flag.String("ref",".gitdb","The dir story Refs")
	argForce    = flag.Bool("f",false,"force re clone")
	argGithub   = flag.String("g","github.com","github server")
)

func main() {
	flag.Parse()
	
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
	for k,v := range rec {
		fmt.Println("Process",k,"stars")
		for num,name := range v {
			repo := strings.Split(name,"/")
			if len(repo)!=2 {
				fmt.Println("error name",name)
				continue
			}
			owner,project := repo[0],repo[1]
			if rdb.OK(owner,project) {
				dump(rdb.GetRef(owner,project))
				continue
			}
			fmt.Println("Begin to Clone",owner,project,num)
			url  := fmt.Sprintf("http://%s/%s/%s",*argGithub,owner,project)	
			path := fmt.Sprintf("%s/repos/%s/%s",*argReposDir,owner,project)
			_, err := os.Stat(path)
			if err == nil {
				if *argForce {
						os.RemoveAll(path)
					} else {
						continue
					}				
			}
			r, err := gitext.PlainCloneFS(url,path)
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
			rdb.PutRef(owner,project,refs)
			dump(refs)
		}
	}
	rdb.Stop()
}

func dump(ref []gitext.Ref) {
	for k,v := range ref {
		fmt.Println(k,":",v)
	}
}