package packext

import (
	"bytes"

	"github.com/a4a881d4/gitcrawling/types"
	"gopkg.in/src-d/go-git.v4/plumbing/format/packfile"
)

type ObjectGet struct {
	g types.PackDataGeter
	s *packfile.Scanner
}

func NewObjectGet(g types.PackDataGeter) *ObjectGet {
	return &ObjectGet{
		g: g,
	}
}

func (og *ObjectGet) Get(oh types.Hash) ([]byte, error) {
	raw, err := og.g.Get(oh)
	if err != nil {
		return []byte{}, err
	}
	og.s = packfile.NewScanner(bytes.NewReader(raw))
	return raw, nil
}

func (og *ObjectGet) Header() (*packfile.ObjectHeader, error) {
	return og.s.SeekObjectHeader(0)
}

func (og *ObjectGet) HeaderByHash(oh types.Hash) (*packfile.ObjectHeader, error) {
	_, err := og.Get(oh)
	if err != nil {
		return nil, err
	}
	return og.Header()
}
