package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/a4a881d4/gitcrawling/db"
	"github.com/a4a881d4/gitcrawling/gitext"
)

var (
	argReposDir = flag.String("r","~/winworks/gitdb","The dir story every thing")
	argForce    = flag.Bool("f",false,"force re clone")
)

func main() {
	flag.Parse()
	var ref []gitext.Ref
	var err error
	if len(flag.Args()) != 1 {
		fmt.Println("Bad args",os.Args)
	}
	repo := strings.Split(flag.Args()[0],"/")
	if len(repo) != 2 {
		fmt.Println("Bad repo",os.Args)
	}
	owner,project := repo[0],repo[1]
	
	var rdb *db.DB
	rdb, err = db.NewDB(*argReposDir+"/refs")
	if err != nil {
		fmt.Println(err)
		return
	}
	var has bool
	if has,err = rdb.HasRef(owner,project); has {
		if ref,err = rdb.GetRef(owner,project); err != nil {
			dump(ref)
		}
	}
	if *argForce {
		url := fmt.Sprintf("http://github.com/%s/%s.git",owner,project)
		ref,err = gitext.Clone(url,*argReposDir)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = rdb.PutRefSync(owner,project,ref)
		dump(ref)
	}
}

func dump(ref []gitext.Ref) {
	for k,v := range ref {
		fmt.Println(k,":",v)
	}
}