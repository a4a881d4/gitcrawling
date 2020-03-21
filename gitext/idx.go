package gitext

import (
	"io"
	"os"

	"gopkg.in/src-d/go-git.v4/plumbing/format/idxfile"
	"gopkg.in/src-d/go-git.v4/plumbing/format/packfile"
)

func NewIdx(idxf string) (*idxfile.MemoryIndex, error) {
	r, err := os.Open(idxf)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	dec := idxfile.NewDecoder(r)
	idx := idxfile.NewMemoryIndex()

	err = dec.Decode(idx)
	return idx, err
}

type Scanner struct {
	packfile.Scanner
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{*packfile.NewScanner(r)}
}

func (s *Scanner) ObjectHeaderAtOffset(offset int64) (*packfile.ObjectHeader, error) {
	return s.SeekObjectHeader(offset)
}
func (s *Scanner) Close() error {
	return s.Scanner.Close()
}
