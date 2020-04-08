package packext

import (
	"bytes"
	"fmt"

	"github.com/a4a881d4/gitcrawling/types"
	"gopkg.in/src-d/go-git.v4/plumbing"
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

func (og *ObjectGet) Get(oh types.Hash) ([]byte, types.Hash, error) {
	raw, base, err := og.g.Get(oh)
	if err != nil {
		return []byte{}, types.ZeroHash, err
	}
	og.s = packfile.NewScanner(bytes.NewReader(raw))
	return raw, base, nil
}

func (og *ObjectGet) Header() (*packfile.ObjectHeader, error) {
	return og.s.SeekObjectHeader(0)
}

func (og *ObjectGet) HeaderByHash(oh types.Hash) (*packfile.ObjectHeader, error) {
	_, base, err := og.Get(oh)
	if err != nil {
		return nil, err
	}
	head, err := og.Header()
	if err != nil {
		return nil, err
	}
	if head.Type.IsDelta() {
		if base != types.ZeroHash {
			head.Type = plumbing.REFDeltaObject
			copy(head.Reference[:], base[:])
		} else {
			return nil, fmt.Errorf("Miss Base")
		}
	}
	return head, nil
}

func (og *ObjectGet) Body(oh types.Hash) ([]byte, error) {
	head, err := og.HeaderByHash(oh)
	if err != nil {
		return []byte{}, err
	}
	buf := new(bytes.Buffer)
	buf.Reset()

	l, _, err := og.s.NextObject(buf)
	if err != nil {
		return []byte{}, err
	}
	if l != head.Length {
		return []byte{}, fmt.Errorf("error length")
	}
	rbuf := buf.Bytes()
	if head.Type.IsDelta() {
		base, err := og.Body(types.Hash(head.Reference))
		if err != nil {
			return []byte{}, err
		}

		patched, err := packfile.PatchDelta(base, rbuf)
		if err != nil {
			return []byte{}, err
		}
		return patched, nil
	}
	return rbuf, nil
}
