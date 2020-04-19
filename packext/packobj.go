package packext

import (
	"bufio"
	"bytes"
	gobinary "encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/a4a881d4/gitcrawling/types"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/format/idxfile"
	"gopkg.in/src-d/go-git.v4/plumbing/format/packfile"
	"gopkg.in/src-d/go-git.v4/utils/binary"
)

type ByteClassify byte

func (b ByteClassify) Hit(h types.Hash) bool {
	return h[0] == byte(b)
}

func (b ByteClassify) NamePrefix() string {
	return fmt.Sprintf("packn-%02x-", b)
}

const (
	isO64Mask = uint64(1) << 31
	noMapping = -1
)

type SplitIdx []types.Classifer

func (s SplitIdx) NamePrefix(k int) (string, error) {
	if k >= len(s) {
		return "", fmt.Errorf("Out of order")
	}
	return s[k].NamePrefix(), nil
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

type IdxObj struct {
	objs        []*ObjEntry
	objm        map[uint64]*ObjEntry
	idxf, packf string
	idx         int
	verbosity   bool
	pf          *PackFile
	seen        map[plumbing.Hash]bool
}

func NewIdxObj(idxf string) (r *IdxObj, err error) {
	r = &IdxObj{}
	r.objm = make(map[uint64]*ObjEntry)
	r.idxf = idxf
	r.packf = strings.Replace(idxf, ".idx", ".pack", -1)
	r.seen = make(map[plumbing.Hash]bool)

	var op OriginPackFile
	op, err = DefaultOPS.GetHash(idxf)
	if err != nil {
		return
	}
	r.pf, err = NewPackFileFromFN(r.packf, types.Hash(op))
	if err != nil {
		return
	}
	var idx *idxfile.MemoryIndex
	idx, err = NewIdx(idxf)
	if err != nil {
		return
	}
	var pfSize uint64
	var stat os.FileInfo
	stat, err = os.Stat(r.packf)
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
		// var dup = &idxfile.Entry{}
		// dup.Hash = last.Hash
		// dup.CRC32 = last.CRC32
		// dup.Offset = last.Offset
		var objentry = op.NewEntry(last)
		objentry.Size = uint32(size)
		r.objs = append(r.objs, objentry)
		r.objm[last.Offset] = objentry
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
	// for k, v := range r.objm {
	// 	fmt.Println(k, v.Offset)
	// }
	return
}
func (idx *IdxObj) Reset(v bool) {
	idx.verbosity = v
	idx.idx = 0
}

func (idx *IdxObj) Close() error {
	return idx.pf.Close()
}
func (idx *IdxObj) GetBase(o *ObjEntry) (*ObjEntry, error) {
	if o.OHeader.Type != plumbing.OFSDeltaObject {
		return nil, fmt.Errorf("has not base %s", o.OHeader.Type.String())
	}
	base, ok := idx.objm[uint64(o.OHeader.OffsetReference)]
	if !ok {
		return nil, fmt.Errorf("Miss base, cannot find base by offset %d", o.OHeader.OffsetReference)
	}
	return base, nil
}

func (idx *IdxObj) ResolveRealType(o *ObjEntry) error {
	if o.RealType != plumbing.InvalidObject {
		if _, ok := idx.seen[o.Hash]; ok {
			return nil
		} else {
			return fmt.Errorf("Miss base, have not seen, but real type set")
		}
	}
	if o.OHeader.Type.IsDelta() {
		base, err := idx.GetBase(o)
		if err != nil {
			return err
		}
		err = idx.ResolveRealType(base)
		if err != nil {
			return err
		}
		o.RealType = base.RealType
	} else {
		o.RealType = o.OHeader.Type
	}
	idx.seen[o.Hash] = true
	return nil
}
func (idx *IdxObj) BuildHead(o *ObjEntry) (head []byte, err error) {
	err = idx.ResolveRealType(o)
	if err != nil {
		return
	}
	head = make([]byte, 5)
	head[0] = byte(o.RealType)
	gobinary.BigEndian.PutUint32(head[1:], uint32(o.OHeader.Length))
	if o.OHeader.Type.IsDelta() {
		head[0] |= 0x10
		var base *ObjEntry
		base, err = idx.GetBase(o)
		if err != nil {
			return
		}
		head = append(head, base.Bytes()...)
	}
	return
}
func (idx *IdxObj) Next(cb func(o *ObjEntry, h, b []byte) error) (err error) {
	if idx.idx == len(idx.objs) {
		err = io.EOF
		return
	}
	o := idx.objs[idx.idx]
	idx.idx++
	var raw, head, body []byte
	raw, err = idx.pf.Get(o)
	if err != nil {
		return
	}
	r := bytes.NewReader(raw)
	s := bufio.NewReader(r)
	var pos int
	o.OHeader, pos, err = processOne(s, o)
	hpos := len(raw) - (pos + r.Len())
	// fmt.Println(o.OHeader.Type.String(), len(raw), pos, r.Len())
	body = raw[hpos:]
	head, err = idx.BuildHead(o)
	if err != nil {
		return err
	}
	o.Size = uint32(len(head) + len(body))
	return cb(o, head, body)
}

func processOne(r *bufio.Reader, o *ObjEntry) (h *packfile.ObjectHeader, pos int, err error) {
	h = &packfile.ObjectHeader{}
	h.Type, h.Length, err = readObjectTypeAndLength(r)
	if err != nil {
		return
	}

	switch h.Type {
	case plumbing.OFSDeltaObject:
		var no int64
		no, err = binary.ReadVariableWidthInt(r)
		if err != nil {
			return
		}

		h.OffsetReference = int64(o.Offset) - no
	case plumbing.REFDeltaObject:
		h.Reference, err = binary.ReadHash(r)
		if err != nil {
			return
		}
	}
	pos = r.Buffered()
	return
}

func readType(r *bufio.Reader) (plumbing.ObjectType, byte, error) {
	var c byte
	var err error
	if c, err = r.ReadByte(); err != nil {
		return plumbing.ObjectType(0), 0, err
	}

	typ := parseType(c)

	return typ, c, nil
}

func parseType(b byte) plumbing.ObjectType {
	return plumbing.ObjectType((b & maskType) >> firstLengthBits)
}

func readObjectTypeAndLength(s *bufio.Reader) (plumbing.ObjectType, int64, error) {
	t, c, err := readType(s)
	if err != nil {
		return t, 0, err
	}

	l, err := readLength(s, c)

	return t, l, err
}

func readLength(r *bufio.Reader, first byte) (int64, error) {
	length := int64(first & maskFirstLength)

	c := first
	shift := firstLengthBits
	var err error
	for c&maskContinue > 0 {
		if c, err = r.ReadByte(); err != nil {
			return 0, err
		}

		length += int64(c&maskLength) << shift
		shift += lengthBits
	}

	return length, nil
}
