package gitext

import (
	"fmt"

	"gopkg.in/src-d/go-billy.v4/osfs"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/storage/filesystem"
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
	return memrefToRef(storage.ReferenceStorage),nil
}

func CloneToFS(path,url string,cb func(*object.Blob) error) ([]Ref,error) {
	
	r, err := PlainCloneFS(url,path)
	if err != nil {
		return nil,err
	}

	ref, err := r.Head()
	if err != nil {
		return nil,err
	}
	fmt.Println("HEAD: ", ref.Hash().String())
	iter,err := r.BlobObjects()
	if err != nil {
		return nil,err
	}
	err = iter.ForEach(cb)
	storage := r.Storer.(*filesystem.Storage)
	return fsrefToRef(storage.ReferenceStorage),nil
}

func memrefToRef(r memory.ReferenceStorage) (ret []Ref) {
	for k,v := range r {
		i := Ref{k.String(),v.Hash().String()}
		ret = append(ret,i)
	}
	return
}

func fsrefToRef(s filesystem.ReferenceStorage) (ret []Ref) {
	iter,err := s.IterReferences()
	if err!=nil {
		return []Ref{}
	}
	iter.ForEach(func(v *plumbing.Reference) error{
		i := Ref{v.Name().String(),v.Hash().String()}
		ret = append(ret,i)
		return nil
	})
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
