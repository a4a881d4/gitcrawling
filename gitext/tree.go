package gitext

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/filemode"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func TreeFlat(r *git.Repository) ([]string, error) {

	ref, err := r.Head()
	if err != nil {
		return []string{}, err
	}

	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		return []string{}, err
	}

	tree, err := r.TreeObject(commit.TreeHash)
	if err != nil {
		return []string{}, err
	}

	seen := make(map[plumbing.Hash]bool)
	walker := object.NewTreeWalker(tree, true, seen)
	var ret = []string{}

	for {
		name, entry, err := walker.Next()

		if err == io.EOF {
			return ret, nil
		}
		if err != nil {
			return ret, err
		}
		ret = append(ret, fmt.Sprintf("%s %s %s", name, entry.Mode.String(), entry.Hash.String()))
	}
}

func Trees(r *git.Repository, cb func(k, v []byte) error) error {
	refs, err := r.Storer.IterReferences()
	if err != nil {
		return err
	}
	seen := make(map[plumbing.Hash]bool)
	var alltree []plumbing.Hash
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		fmt.Println(ref.String())
		if ref.Hash().IsZero() {
			return nil
		}
		commit, err := r.CommitObject(ref.Hash())
		if err != nil {
			return err
		}

		alltree = append(alltree, commit.TreeHash)
		tree, err := r.TreeObject(commit.TreeHash)
		if err != nil {
			return err
		}

		walker := object.NewTreeWalker(tree, true, seen)
		for {
			_, entry, err := walker.Next()

			if err == io.EOF {
				return nil
			}

			if entry.Mode == filemode.Dir {
				alltree = append(alltree, entry.Hash)
			}
		}
	})
	if err != nil {
		return err
	}

	for _, h := range alltree {
		o, err := r.Storer.EncodedObject(plumbing.TreeObject, h)
		if err != nil {
			return err
		}
		b, err := EncodedTree(o)

		if err != nil {
			return err
		}

		err = cb(h[:], b)
		if err != nil {
			return err
		}
	}
	return nil
}

func DumpTree(b []byte) ([]*object.TreeEntry, error) {
	r := bufio.NewReader(bytes.NewBuffer(b))
	ret := []*object.TreeEntry{}
	for {
		str, err := r.ReadString(' ')
		if err != nil {
			if err == io.EOF {
				break
			}

			return ret, err
		}
		str = str[:len(str)-1] // strip last byte (' ')

		mode, err := filemode.New(str)
		if err != nil {
			return ret, err
		}

		name, err := r.ReadString(0)
		if err != nil && err != io.EOF {
			return ret, err
		}

		var hash plumbing.Hash
		if _, err = io.ReadFull(r, hash[:]); err != nil {
			return ret, err
		}

		baseName := name[:len(name)-1]
		ret = append(ret, &object.TreeEntry{
			Hash: hash,
			Mode: mode,
			Name: baseName,
		})
	}
	return ret, nil
}

func TreeEntry2String(e *object.TreeEntry) string {
	return e.Hash.String() + ": " + e.Mode.String() + " " + e.Name
}

type Inode struct {
	Mode filemode.FileMode
	Name string
}

func NewInode(e *object.TreeEntry) *Inode {
	return &Inode{e.Mode, e.Name}
}

func Entry2Map(entries []*object.TreeEntry, m map[plumbing.Hash]*Inode) map[plumbing.Hash]*Inode {
	for _, e := range entries {
		m[e.Hash] = NewInode(e)
	}
	return m
}

type DBGeter interface {
	GetRawTree(k []byte, cb func(v []byte) error) error
}

func done(h plumbing.Hash, g DBGeter, m map[plumbing.Hash]*Inode) error {
	err := g.GetRawTree(h[:], func(v []byte) error {
		entries, err := DumpTree(v)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if !e.Mode.IsFile() {
				err := done(e.Hash, g, m)
				if err != nil {
					return err
				}
			}
		}
		Entry2Map(entries, m)
		return nil
	})
	return err
}
func Tree(root plumbing.Hash, g DBGeter) (map[plumbing.Hash]*Inode, error) {
	m := make(map[plumbing.Hash]*Inode)
	err := done(root, g, m)
	return m, err
}
