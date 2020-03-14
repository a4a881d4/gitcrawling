package gitext

import (
	"fmt"
	"io"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
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
