package main

import (
	"fmt"
	"io"
	"os"

	"github.com/a4a881d4/gitcrawling/gitext"
	. "github.com/a4a881d4/gitcrawling/ref"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

// Basic example of how to clone a repository using clone options.
func main() {
	CheckArgs("<directory>")
	directory := os.Args[1]

	r, err := git.PlainOpen(directory)
	CheckIfError(err)

	all, err := gitext.TreeFlat(r)
	if err != nil {
		fmt.Println(err)
	} else {
		for k, v := range all {
			fmt.Println(format(k, v))
		}
	}
	ref, err := r.Head()
	CheckIfError(err)
	// ... retrieving the commit object
	commit, err := r.CommitObject(ref.Hash())
	CheckIfError(err)

	fmt.Println(commit)

	tree, err := r.TreeObject(commit.TreeHash)
	CheckIfError(err)

	seen := make(map[plumbing.Hash]bool)
	walker := object.NewTreeWalker(tree, true, seen)

	for {
		name, entry, err := walker.Next()
		// if _, ok := seen[entry.Hash]; ok {
		// 	continue
		// }
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Printf("%s: %s %s\n", entry.Hash.String(), entry.Mode.String(), name)
	}
	fmt.Println(commit.TreeHash.String())
}

func format(k int, v string) string {
	return fmt.Sprintf("%5d ", k) + v
}
