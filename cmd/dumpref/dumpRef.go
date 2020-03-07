package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

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
				fmt.Println(owner,project,"miss")
			}
			if rdb.IsBuild(owner,project) {
				fmt.Println(owner,project,"has build")
			}
			if rdb.OK(owner,project) {
				fmt.Println(owner,project,"clone local")
			}
		}
	}
}
