package main

import (
	"fmt"
	"os"

	. "github.com/a4a881d4/gitcrawling/ref"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
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
}
