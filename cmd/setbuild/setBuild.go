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
	argReposDir = flag.String("r",".gitdb","The dir story every thing")
)

func main() {
	flag.Parse()
	var ref []gitext.Ref
	var err error

	rdb := db.NewRefDB(*argReposDir+"/refs")
	
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
				fmt.Println("miss",owner,project)
				continue
			}
			if rdb.IsBuild(owner,project) {
				fmt.Println("has build",owner,project)
			}
			fmt.Println("Set build object flag",owner,project)
			
			ref,err = rdb.SetBuild(owner,project,[]gitext.Ref{})
			if err != nil {
				fmt.Println(err)
			} else {
				dump(ref)
			}
		}
	}
}

func dump(ref []gitext.Ref) {
	for k,v := range ref {
		fmt.Println(k,":",v)
	}
}