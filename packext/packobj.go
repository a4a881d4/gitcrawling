package packext

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/a4a881d4/gitcrawling/types"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/format/idxfile"
)

type ByteClassify byte

func (b ByteClassify) Hit(h types.Hash) bool {
	return h[0] == byte(b)
}

func (b ByteClassify) FileNamePrefix() string {
	return fmt.Sprintf("packn-%02x-", b)
}

const (
	isO64Mask = uint64(1) << 31
	noMapping = -1
)

type SplitIdx []types.Classifer

func (s SplitIdx) FileNamePrefix(k int) (string, error) {
	if k >= len(s) {
		return "", fmt.Errorf("Out of order")
	}
	return s[k].FileNamePrefix(), nil
}

func (s SplitIdx) Number(hash plumbing.Hash) (hs []int) {
	for k, v := range s {
		if v.Hit(types.Hash(hash)) {
			hs = append(hs, k)
		}
	}
	return
}

func DefaultByteSplit() SplitIdx {
	r := make([]types.Classifer, 256)
	for k, _ := range r {
		r[k] = ByteClassify(k)
	}
	return r
}

func (s SplitIdx) GetOffset(idxf string) (objs [][]ObjEntry, err error) {
	var op OriginPackFile
	op, err = DefaultOPS.GetHash(idxf)
	if err != nil {
		return
	}
	objs = make([][]ObjEntry, len(s))

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
		if len(ma) == 0 {
			return
		}
		var objentry = op.NewEntry(last)
		objentry.Size = uint32(size)
		for _, mma := range ma {
			objs[mma] = append(objs[mma], *objentry)
		}
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
