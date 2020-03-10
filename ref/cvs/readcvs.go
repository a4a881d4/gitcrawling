package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/a4a881d4/gitcrawling/db"
	"github.com/a4a881d4/gitcrawling/gitext"
)

func main() {
	// db, err := db.NewDB(os.Args[2])
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// defer db.Close()

	for k:=1;k<67;k++ {
		fn := fmt.Sprintf("%s/repo_%d.csv",os.Args[1],k)
		ProcessCVS(fn,db)
	}
}

func ProcessCVS(fn string, db *db.DB) {
	f, err := os.Open(fn)
	if err != nil {
		fmt.Println(err)
		return
	}
	r := csv.NewReader(f)
	Items, err := r.ReadAll()
	if err != nil {
		fmt.Println(err)
		return
	}
	for k, v := range Items {
		owner := v[2]
		project := v[1]
		star, _ := strconv.ParseInt(v[4], 10, 64)
		refs := gitext.EmptyRef()
		
		r := gitext.Record{
			v[5],
			uint64(star),
			refs,
			uint64(time.Now().Unix()),
		}
		fmt.Println(k, owner, project)
		fmt.Println(r.String())
		// if k != 0 {
		// 	if err := db.PutRepo(owner, project, &r); err != nil {
		// 		fmt.Println(err)
		// 	}
		// }
	}
}