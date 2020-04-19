package packext

import "github.com/a4a881d4/gitcrawling/types"

type HeaderExt [32]byte

type Selector interface {
	Determine([]*ObjEntry) *ObjEntry
}

type maxSelect struct{}
type MatchSelect types.Hash

func (maxSelect) Determine(objs []*ObjEntry) (obj *ObjEntry) {
	if len(objs) == 0 {
		return nil
	}
	var size uint32
	for _, v := range objs {
		cmp := NewOEFromBytes(v[:])
		if cmp.Size > size {
			obj = cmp
		}
	}
	return
}

func (hash MatchSelect) Determine(objs []HeaderExt) *ObjEntry {
	if len(objs) == 0 {
		return nil
	}
	var maxidx = 0
	for k, v := range objs {
		if v.Size > objs[maxidx].Size {
			maxidx = k
		}
	}
	return objs[maxidx]
}

var MaxSelect maxSelect
