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
)

var (
	argReposDir = flag.String("r",".gitdb","The dir story every thing")
)

func main() {
	flag.Parse()
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
			}
			if rdb.IsBuild(owner,project) {
				fmt.Println(ShowName(owner,project),"has build")
			}
			if rdb.OK(owner,project) {
				fmt.Println(ShowName(owner,project),"clone local")
			}
			<- time.After(time.Second*1)
		}
	}
}

func ShowName(owner,project string) string {
	var space = "                                                                        "
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
	return owner+":"+project
}