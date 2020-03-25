package gitext

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/format/idxfile"
	"gopkg.in/src-d/go-git.v4/utils/binary"
)

type ObjEntry struct {
	Offset, Size uint64
}

type Classifer interface {
	Hit(plumbing.Hash) bool
}

func (b byte) Hit(h plumbing.Hash) bool {
	h[0] == b
}

const (
	isO64Mask = uint64(1) << 31
	noMapping = -1
)

// func getOffset(idx *idxfile.MemoryIndex, v []byte) uint64 {
// 	ofs := encbin.BigEndian.Uint32(v[:])

// 	if (uint64(ofs) & isO64Mask) != 0 {
// 		offset := 8 * (uint64(ofs) & ^isO64Mask)
// 		n := encbin.BigEndian.Uint64(idx.Offset64[offset : offset+8])
// 		return n
// 	}

// 	return uint64(ofs)
// }

type SplitIdx []Classifer

func (s SplitIdx) GetOffset(idxf string) (objs [][]ObjEntry, err error) {
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
	offmap := make(map[uint64]uint64)

	iter, err := idx.EntriesByOffset()
	if err != nil {
		return
	}
	var last, e *idxfile.Entry
	last = nil
	for {
		e, err = iter.Next()
		fmt.Println("e:", e, last)
		if err == io.EOF {
			offmap[last.Offset] = pfSize - 20 - last.Offset
			err = nil
			break
		}
		if err != nil {
			return
		}
		if last != nil {
			offmap[last.Offset] = e.Offset - last.Offset
		}
		last = e
	}
	for l := 0; l < 256; l++ {
		i := idx.FanoutMapping[l]
		if i == noMapping {
			continue
		}
		num := len(idx.Offset32[i]) >> 2
		for k := 0; k < num; k++ {
			offset := getOffset(idx, idx.Offset32[i][k*4:(k+1)*4])
			objs[l] = append(objs[l], ObjEntry{offset, offmap[offset]})
		}
	}
	return
}

type PackFile struct {
	f      os.File
	w      io.Writer
	hasher plumbing.Hasher
}

func NewPack(fn string) (*PackFile, error) {
	w, err := os.Create(fn)
	if err != nil {
		return nil, err
	}
	h := plumbing.Hasher{
		Hash: sha1.New(),
	}
	mw := io.MultiWriter(w, h)
	return &PackFile{
		f:      w,
		w:      mw,
		hasher: h,
	}
}

var signature = []byte{'P', 'A', 'C', 'K'}

const (
	// VersionSupported is the packfile version supported by this package
	VersionSupported uint32 = 2
)

func (pf *PackFile) Head(numEntries int) error {
	return binary.Write(
		pf.w,
		signature,
		int32(VersionSupported),
		int32(numEntries),
	)
}

func (pf *PackFile) Footer() (plumbing.Hash, error) {
	h := pf.hasher.Sum()
	return h, binary.Write(pf.w, h)
}

func (pf *PackFile) Do(r io.Reader, size int64) (int64, error) {
	return io.CopyN(pf.w, r, size)
}

func (pf *PackFile) Close() error {
	pf.f.Close()
}
