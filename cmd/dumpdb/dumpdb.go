package main

import (
	"fmt"
	"os"

	"github.com/a4a881d4/gitcrawling/db"
	"github.com/a4a881d4/gitcrawling/gitext"
)

func main() {
	db, err := db.NewDB(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()
	api := gitext.NewGitHubClient()
	db.ForEachRepo(func(owner, project string, r *gitext.Record) {
		fmt.Println(owner, project)
		fmt.Println(r.String())
		if len(r.Refs)==0 {
			refs,err := api.GetRef(owner,project)
			if err != nil {
				fmt.Println(err)
				return
			}
			if len(refs)>0 {
				r.Refs = refs
				db.PutRepo(owner, project,r)
			}
		}
	})
}
