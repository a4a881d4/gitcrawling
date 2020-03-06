package gitext

import (
	"fmt"

	"gopkg.in/src-d/go-billy.v4/osfs"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/storage/filesystem/dotgit"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

func Clone(url,repos string) ([]Ref,error) {
	
	r, err := PlainClone(url)
	if err != nil {
		return nil,err
	}

	ref, err := r.Head()
	if err != nil {
		return nil,err
	}
	fmt.Println("HEAD: ", ref.Hash().String())

	storage := r.Storer.(*memory.Storage)

	gfs := osfs.New(repos)
	dir := dotgit.New(gfs)
	for k, o := range storage.Objects {
		if _, err := SetEncodedObject(dir, o); err != nil {
			fmt.Println(k, err)
		}
	}
	return refToRef(storage.ReferenceStorage),nil
}

func refToRef(r memory.ReferenceStorage) (ret []Ref) {
	for k,v := range r {
		i := Ref{k.String(),v.Hash().String()}
		ret = append(ret,i)
	}
	return
}

func CloneToMem(url string) ([]Ref,map[plumbing.Hash][]byte,error) {
	blobs := make(map[plumbing.Hash][]byte)
	r, err := PlainClone(url)
	if err != nil {
		return nil,map[plumbing.Hash][]byte{},err
	}

	ref, err := r.Head()
	if err != nil {
		return nil,map[plumbing.Hash][]byte{},err
	}
	fmt.Println("HEAD: ", ref.Hash().String())

	storage := r.Storer.(*memory.Storage)

	for k, o := range storage.Objects {
		if hash, blob, err := EncodedObject(o); err != nil {
			fmt.Println(k, err)
		} else {
			blobs[hash]=blob
		}
	}
	return refToRef(storage.ReferenceStorage),blobs,nil
}
