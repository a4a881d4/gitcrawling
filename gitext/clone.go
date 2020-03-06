package gitext

import (
	"fmt"

	"gopkg.in/src-d/go-billy.v4/osfs"
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

	gfs := osfs.New(repos+"/objects")
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

