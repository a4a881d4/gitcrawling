package packext

import (
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/format/idxfile"
	"gopkg.in/src-d/go-git.v4/plumbing/format/packfile"
)

func GetOffsetNoClassify(idxf string) (op OriginPackFile, objs []ObjEntry, err error) {
	op, err = DefaultOPS.GetHash(idxf)
	if err != nil {
		return
	}
	objs = []ObjEntry{}

	var idx *idxfile.MemoryIndex
	idx, err = NewIdx(idxf)
	if err != nil {
		return
	}
	var pfSize uint64
	packf := strings.Replace(idxf, ".idx", ".pack", -1)
	stat, err := os.Stat(packf)
	if err != nil {
		return
	}
	pfSize = uint64(stat.Size())
	pf, err := os.Open(packf)
	if err != nil {
		return
	}
	defer pf.Close()
	scanner := packfile.NewScanner(pf)
	iter, err := idx.EntriesByOffset()
	if err != nil {
		return
	}
	var last, e *idxfile.Entry
	var mobjs = make(map[uint64]*ObjEntry)
	last = nil
	var Add = func(size uint64) {
		if last == nil {
			return
		}

		var objentry = op.NewEntry(last)
		var errs error
		objentry.OHeader, errs = scanner.SeekObjectHeader(int64(last.Offset))
		if errs != nil {
			fmt.Println(errs)
		}
		objentry.Size = uint32(size)
		objs = append(objs, *objentry)
		mobjs[last.Offset] = objentry
	}

	for {
		e, err = iter.Next()
		if err == io.EOF {
			Add(pfSize - 20 - last.Offset)
			err = nil
			break
		}
		if err != nil {
			return
		}
		if last != nil {
			Add(e.Offset - last.Offset)
		}
		last = e
	}
	var getRealType func(uint64) plumbing.ObjectType
	getRealType = func(off uint64) plumbing.ObjectType {
		if base, ok := mobjs[off]; ok {
			if base.OHeader.Type.IsDelta() {
				return getRealType(uint64(base.OHeader.OffsetReference))
			} else {
				return base.OHeader.Type
			}
		}
		return plumbing.AnyObject
	}
	for k, e := range objs {
		if e.OHeader.Type == plumbing.OFSDeltaObject {
			if base, ok := mobjs[uint64(e.OHeader.OffsetReference)]; ok {
				copy(e.OHeader.Reference[:], base.Hash[:])
				objs[k].RealType = getRealType(uint64(e.OHeader.OffsetReference))
			} else {
				fmt.Println("Miss base object of a OFSDeltaObject", e.OHeader.Offset, e.OHeader.OffsetReference)
			}
		} else {
			objs[k].RealType = e.OHeader.Type
		}
	}
	return
}
