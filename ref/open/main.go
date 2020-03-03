package main

import (
	"fmt"
	"os"
	"io"

	. "github.com/a4a881d4/gitcrawling/ref"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// Basic example of how to clone a repository using clone options.
func main() {
	CheckArgs("<directory>")
	directory := os.Args[1]

	r, err := git.PlainOpen(directory)
	CheckIfError(err)

	it,_ := r.TreeObjects()
	it.ForEach(func(t *object.Tree) error{
		for k,v := range t.Entries {
			fmt.Println(k,v.Name,v.Mode)
		}
		fmt.Println(t.Type())
		return nil
		})

	ref, err := r.Head()
	CheckIfError(err)

	tree,err := r.TreeObject(ref.Hash())
	CheckIfError(err)

	seen := make(map[plumbing.Hash]bool)
	walker := object.NewTreeWalker(tree,false,seen)

	for {
		name, entry, err := walker.Next()
		if _,ok := seen[entry.Hash]; ok {
			continue
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Printf("name:%s hash:%x mode:%s\n",name,entry.Hash,entry.Mode.String())
	}
}
