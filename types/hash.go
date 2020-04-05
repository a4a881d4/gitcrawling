package types

import "gopkg.in/src-d/go-git.v4/plumbing"

type Hash plumbing.Hash

func(h Hash) String() string {
	return plumbing.Hash(h).String()
}

type PackDataGeter interface {
	Get(Hash) []byte
} 