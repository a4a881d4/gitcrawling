package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/a4a881d4/gitcrawling/gitext"
)

var (
	argReposDir = flag.String("r",".gitdb","The dir story every thing")
	argForce    = flag.Bool("f",false,"force re clone")
	argGithub   = flag.String("g","github.com","github server")
)

func main() {
	flag.Parse()
	var ref []gitext.Ref
	var err error

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
			url  := fmt.Sprintf("http://%s/%s/%s",*argGithub,owner,project)	
			path := fmt.Sprintf("%s/repos/%s/%s",*argReposDir,owner,project)
			r, err := gitext.PlainCloneFS(url,path)
			if err != nil {
				fmt.Println(err)
			}
		
			ref, err := r.Head()
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("HEAD: ", ref.Hash().String())
		}
	}
}

