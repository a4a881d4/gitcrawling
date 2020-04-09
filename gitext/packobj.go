package gitext

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/a4a881d4/gitcrawling/packext"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/format/idxfile"
	"gopkg.in/src-d/go-git.v4/plumbing/format/packfile"
)

type Classifer interface {
	Hit(plumbing.Hash) bool
	FileNamePrefix() string
}
type ByteClassify byte

func (b ByteClassify) Hit(h plumbing.Hash) bool {
	return h[0] == byte(b)
}

func (b ByteClassify) FileNamePrefix() string {
	return fmt.Sprintf("pack-%02x-", b)
}

const (
	isO64Mask = uint64(1) << 31
	noMapping = -1
)

type SplitIdx []Classifer

func (s SplitIdx) FileNamePrefix(k int) (string, error) {
	if k >= len(s) {
		return "", fmt.Errorf("Out of order")
	}
	return s[k].FileNamePrefix(), nil
}

func (s SplitIdx) Number(hash plumbing.Hash) int {
	for k, v := range s {
		if v.Hit(hash) {
			return k
		}
	}
	return -1
}

func DefaultByteSplit() SplitIdx {
	r := make([]Classifer, 256)
	for k, _ := range r {
		r[k] = ByteClassify(k)
	}
	return r
}

func (s SplitIdx) GetOffset(idxf string) (objs [][]packext.ObjEntry, err error) {
	var op packext.OriginPackFile
	op, err = packext.DefaultOPS.GetHash(idxf)
	if err != nil {
		return
	}
	objs = make([][]packext.ObjEntry, len(s))

	var idx *idxfile.MemoryIndex
	idx, err = NewIdx(idxf)
	if err != nil {
		return
	}
	var pfSize uint64
	stat, err := os.Stat(strings.Replace(idxf, ".idx", ".pack", -1))
	if err != nil {
		return
	}
	pfSize = uint64(stat.Size())

	iter, err := idx.EntriesByOffset()
	if err != nil {
		return
	}
	var last, e *idxfile.Entry
	last = nil
	var Add = func(size uint64) {
		if last == nil {
			return
		}

		ma := s.Number(last.Hash)
		if ma == -1 {
			return
		}
		var objentry = op.NewEntry(last)
		objentry.Size = uint32(size)
		objs[ma] = append(objs[ma], *objentry)
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

	return
}

func GetOffsetNoClassify(idxf string) (op packext.OriginPackFile, objs []packext.ObjEntry, err error) {
	op, err = packext.DefaultOPS.GetHash(idxf)
	if err != nil {
		return
	}
	objs = []packext.ObjEntry{}

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
	var mobjs = make(map[uint64]*packext.ObjEntry)
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
