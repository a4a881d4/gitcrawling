package main

import (
	"fmt"
	"os"

	"github.com/a4a881d4/gitcrawling/gitext"
	. "github.com/a4a881d4/gitcrawling/ref"
	"gopkg.in/src-d/go-billy.v4/osfs"
	"gopkg.in/src-d/go-git.v4/storage/filesystem/dotgit"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

// Basic example of how to clone a repository using clone options.
func main() {
	CheckArgs("<url>")
	url := os.Args[1]

	// Clone the given repository to the given directory
	// Info("git clone %s %s --recursive", url, directory)

	r, err := gitext.PlainClone(url)

	CheckIfError(err)

	// ... retrieving the branch being pointed by HEAD
	ref, err := r.Head()

	CheckIfError(err)
	fmt.Println("HEAD: ", ref.Hash().String())

	storage := r.Storer.(*memory.Storage)

	gfs := osfs.New(os.Args[2])
	dir := dotgit.New(gfs)
	for _, o := range storage.Objects {
		if _, err := gitext.SetEncodedObject(dir, o); err != nil {
			fmt.Println(0, err)
		}
	}
}
