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
	CheckArgs("<url>", "<directory>")
	url := os.Args[1]
	directory := os.Args[2]

	// Clone the given repository to the given directory
	Info("git clone %s %s --recursive", url, directory)

	r, err := git.PlainClone(directory, false, &git.CloneOptions{
		URL:               url,
		Depth:             1,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Progress:          os.Stdout,
	})

	CheckIfError(err)

	// ... retrieving the branch being pointed by HEAD
	ref, err := r.Head()
	CheckIfError(err)
	// ... retrieving the commit object
	commit, err := r.CommitObject(ref.Hash())
	CheckIfError(err)

	fmt.Println(commit)

	it, _ := r.TreeObjects()
	it.ForEach(func(t *object.Tree) error {
		for k, v := range t.Entries {
			fmt.Println(k, v.Name)
		}
		fmt.Println(t.Type())
		return nil
	})
}
