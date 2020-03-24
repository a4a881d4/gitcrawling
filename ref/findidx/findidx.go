package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/a4a881d4/gitcrawling/gitext"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

type Offset []uint64

func (l Offset) Len() int {
	return len(l)
}
func (l Offset) Less(i, j int) bool {
	return l[i] < l[j]
}
func (l Offset) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
func main() {
	var rec = make([]int, 1024)
	filepath.Walk(os.Args[1], func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, ".idx") {
			fmt.Println(path)
			idx, err := gitext.NewIdx(path)
			if err != nil {
				fmt.Println(2, err)
			}

			iter, err := idx.Entries()
			if err != nil {
				fmt.Println(3, err)
			}
			var offset Offset
			for {
				e, err := iter.Next()
				if err == io.EOF {
					break
				}

				offset = append(offset, e.Offset)
			}
			sort.Sort(offset)
			if len(offset) > 2 {
				for i := 1; i < len(offset); i++ {
					d := offset[i] - offset[i-1]
					d /= 256
					if d > 1023 {
						d = 1023
					}
					rec[d]++
				}
			}

			packfile := strings.Replace(path, ".idx", ".pack", -1)
			st,err := os.Stat(packfile)
			
			if err != nil {
				return err
			}
			pflen  := st.Size()
			r, err := os.Open(packfile)
			if err != nil {
				return err
			}
			defer r.Close()
			s := gitext.NewScanner(r)
			defer s.Close()
			for k, v := range offset {
				h, err := s.SeekObjectHeader(int64(v))
				if err != nil {
					fmt.Println(err)
				} else {
					var length int64
					if k < len(offset)-1 {
						length = int64(offset[k+1] - offset[k])
					} else {
						length = pflen - int64(offset[k]) - 20
					}
					fmt.Printf("%10s: %6d %6d %6d\n", h.Type.String(),offset[k], h.Length, length)
				}
			}
			var hash plumbing.Hash
			hl,err := r.ReadAt(hash[:],pflen-20)
			if hl!=20 || err!=nil {
				fmt.Println(err,hl)
				return err
			} else {
				fmt.Println(hash.String())
			}
		}
		return err
	})
	// sum := 0
	// for k, v := range rec {
	// 	fmt.Printf("%6d ", v)
	// 	if k%8 == 7 {
	// 		fmt.Println()
	// 	}
	// 	sum += v
	// }
	// fmt.Println()
	// fmt.Println(sum)
}
