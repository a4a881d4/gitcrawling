package packext

import (
	"sort"

	"github.com/a4a881d4/gitcrawling/types"
)

type PackFileAndDB struct {
	files *PackFiles
	db    types.DBer
}

func NewFileDirPFDB(db types.DBer, dir string) (*PackFileAndDB, error) {
	err := DefaultFromDir(dir)
	if err != nil {
		return nil, err
	}
	ps := NewPackFiles(nil)
	return &PackFileAndDB{ps, db}, nil
}

func (p *PackFileAndDB) Get(h types.Hash) ([]byte, types.Hash, error) {
	hs := p.db.NewHashGeter(h.String())
	defer hs.End()
	var os Entries
	_, err := hs.NextGroup(45, func() types.Byter {
		po := &ObjEntry{}
		os = append(os, po)
		return po
	})
	if err != nil {
		return []byte{}, types.ZeroHash, err
	}
	sort.Sort(os)
	raw, err := p.files.Get(os[0])
	if err != nil {
		return []byte{}, types.ZeroHash, err
	}
	if os[0].OHeader.Type.IsDelta() {
		return raw, types.Hash(os[0].OHeader.Reference), nil
	}
	return raw, types.ZeroHash, nil
}
