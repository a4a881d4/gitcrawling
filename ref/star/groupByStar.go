package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

var (
	forked = NewStar(20).Init()
	origin = NewStar(20).Init()
)

type StarRecord struct {
	Size int
	Record []map[int64][]string
}

func NewStar(s int) *StarRecord {
	return &StarRecord{Size:s}
}

func(r *StarRecord) Init() *StarRecord {
	for i := 0;i<r.Size;i++ {
		r.Record = append(r.Record,make(map[int64][]string))
	}
	return r
}

func main() {

	for k:=0;k<67;k++ {
		fn := fmt.Sprintf("%s/repo_%d.csv",os.Args[1],k)
		ProcessCVS(fn)
	}
	for k,v := range origin.Record {
		if len(v)>0 {
			fn := fmt.Sprintf("%s/origin%06d.star",os.Args[2],k)
			writeToFile(fn,v)
		}
	}
	for k,v := range forked.Record {
		if len(v)>0 {
			fn := fmt.Sprintf("%s/forked%06d.star",os.Args[2],k)
			writeToFile(fn,v)
		}
	}
}

func writeToFile(fn string,v map[int64][]string) {
	file, err := os.OpenFile(fn,  os.O_RDWR | os.O_APPEND | os. O_CREATE,066)
	
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close() 
	
	ow := bufio.NewWriter(file)
	encoder := json.NewEncoder(ow)
	encoder.Encode(v)
	ow.Flush()
	
}

func order(a int64) int {
	for r:=0;r<20;r++ {
		if a==0 {
			return r
		}
		a = a>>1
	}
	return 19
}

func ProcessCVS(fn string) {
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
		owner    := v[2]
		project  := v[1]
		fullname := owner+"/"+project
		star, err  := strconv.ParseInt(v[4], 10, 64)
		if err != nil {
			continue
		}
		o := order(star)
		if v[3]=="0" {
			origin.Record[o][star] = append(origin.Record[o][star],fullname)
		} else {
			forked.Record[o][star] = append(forked.Record[o][star],fullname)
		}
		if k%100==0 {
			fmt.Printf("*")
		}
	}
	fmt.Println()
}