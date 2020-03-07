package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/a4a881d4/gitcrawling/db"
	"github.com/a4a881d4/gitcrawling/gitext"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	gitutil "gopkg.in/src-d/go-git.v4/utils/ioutil"
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

	var rdb,bdb *db.DB
	rdb, err = db.NewDB(*argReposDir+"/refs")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer rdb.Close()
	bdb, err = db.NewDB(*argReposDir+"/objects")
	if err != nil {
		fmt.Println(err)
		return
	}
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
			if has,err = rdb.HasRef(owner,project); has {
				if ref,err = rdb.GetRef(owner,project); err == nil {
					dump(ref)
				} else {
					fmt.Println(err)
				}
			}
		
			if !has || *argForce {
				fmt.Println("Begin to Clone",owner,project)
				if ref,err = CloneAndSave(owner,project,*argReposDir,rdb,bdb); err == nil {
					dump(ref)
				} else {
					fmt.Println(err)
				}
			}
		}
	}

}

func CloneAndSave(owner,project,ReposDir string, rdb,bdb *db.DB) (ref []gitext.Ref,err error) {
	url  := fmt.Sprintf("http://%s/%s/%s",*argGithub,owner,project)
	path := fmt.Sprintf("%s/repos/%s/%s",ReposDir,owner,project)
	var c = 0
	ref,err = gitext.CloneToFS(path,url,func(b *object.Blob) (ierr error){
		k := b.ID()
		var has bool
		if has,ierr = bdb.HasBlob(k); has || ierr!=nil{
			if has {
				ierr = fmt.Errorf("dup blob %s",k.String())
				fmt.Printf("-")
			}
			return
		}
		
		r,ierr := b.Reader()
		if err != nil {
			fmt.Printf("/")
			return 
		}
		defer gitutil.CheckClose(r,&ierr)
		buf := bytes.NewBuffer(make([]byte,b.Size))
		s,ierr := io.Copy(buf,r)
		if int64(s)!=b.Size {
			fmt.Printf(`?`)
			ierr = fmt.Errorf("blob is too big")
			return
		}
		if ierr != nil {
			fmt.Printf(`s`)
			return
		}
		if ierr = bdb.PutBlob(k,buf.Bytes()); ierr !=nil {
			fmt.Printf(`v`)
			return
		}
		c++
		if c%1000 == 999 {
			fmt.Printf("*")
			os.Stdout.Sync()
		}
		return
	})
	
	fmt.Println("")
	if err != nil {
		fmt.Println(err)
	}
	err = rdb.PutRefSync(owner,project,ref)
	return
}

func dump(ref []gitext.Ref) {
	for k,v := range ref {
		fmt.Println(k,":",v)
	}
}