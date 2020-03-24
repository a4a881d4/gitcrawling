package gitext

import (
	"io"
	"os"
	"path"
	"gopkg.in/src-d/go-git.v4/plumbing"
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

func Reidx(packf string) (hs string,err error) {
	var r,w *os.File
	hs = ""
	r, err = os.Open(packf)
	if err != nil {
		return 
	}
	defer r.Close()

	var parser *packfile.Parser
	
	scanner    := packfile.NewScanner(r)
	idxw       := &idxfile.Writer{}
	parser,err  = packfile.NewParser(scanner,idxw)
	if err!= nil {
		return 
	}
	var h plumbing.Hash
	h, err = parser.Parse()
	if err!= nil {
		return 
	}
	hs = h.String()

	var index *idxfile.MemoryIndex
	index,err = idxw.Index()
	if err!= nil {
		return 
	}
	
	dir := path.Dir(packf)
	w, err = os.Create(path.Join(dir,"pack-"+hs+".idx"))
	if err != nil {
		return
	}
	defer w.Close()

	_,err = idxfile.NewEncoder(w).Encode(index)
	return
}