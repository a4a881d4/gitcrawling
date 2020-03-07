package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/a4a881d4/gitcrawling/db"
	"github.com/a4a881d4/gitcrawling/gitext"
)

var (
	argReposDir = flag.String("r",".gitdb","The dir story every thing")
	argForce    = flag.Bool("f",false,"force re clone")
	argV        = flag.Bool("v",false,"force re clone")
)

func main() {
	flag.Parse()
	var ref []gitext.Ref
	var err error

	rdb := db.NewRefDB(*argReposDir+"/refs")
	
	db, err := db.NewDB(*argReposDir+"/objects")
	if err != nil {
		fmt.Println(err)
		return
	}
	bdb := NewObjectDB(db)
	defer bdb.Close()

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
			var has bool
			if !rdb.OK(owner,project) {
				fmt.Println("miss",owner,project)
				continue
			}
			if rdb.IsBuild(owner,project) {
				fmt.Println("has build",owner,project)
				if !*argForce {
					continue
				}
			}
			fmt.Println("Begin to build objects",owner,project)
			if ref,err = OpenAndSave(owner,project,*argReposDir,bdb); err != nil {
				fmt.Println(err)
			} else {
				dump(ref)
			}
		}
	}
}

func OpenAndSave(owner,project,ReposDir string, bdb *db.ObjDB) (ref []gitext.Ref,err error) {
	path  := fmt.Sprintf("%s/repos/%s/%s",ReposDir,owner,project)
	r,err := gitext.PlainOpen(path)
	if err != nil {
		return []gitext.Ref{},err
	}
	blobs,err := r.BlobObjects()
	if err != nil {
		return []gitext.Ref{},err
	}
	go func(){
		err := bdb.PutObjects(blobs)
		if err != nil {
			fmt.Println(err)
		}
	}()
	bdb.Wait(*argV,100)
	return gitext.RepoRef(r),err
}

func dump(ref []gitext.Ref) {
	for k,v := range ref {
		fmt.Println(k,":",v)
	}
}