package types

import (
	"encoding/binary"
	"fmt"
	"io"
)

type Index uint64
type Indexes []Index
type KeyPart [4]byte
type OffPart uint32

func (ids Indexes) ToFile(w io.Writer) error {
	b := make([]byte, 8)
	for k, i := range ids {
		binary.BigEndian.PutUint64(b, uint64(i))
		n, err := w.Write(b)
		if n != 8 {
			return fmt.Errorf("Write index %d(%d,%d)", k, 8, n)
		}
		if err != nil {
			return fmt.Errorf("Write index %d(%v)", k, err)
		}
	}
	return nil
}
func IndexFromFile(r io.Reader) (ids Indexes, err error) {
	b := make([]byte, 8)
	var n, k int
	n, err = r.Read(b)
	for err == nil {
		if n != 8 {
			err = fmt.Errorf("read index %d(%d,%d)", k, 8, n)
			return
		}
		ids = append(ids, Index(binary.BigEndian.Uint64(b)))
		n, err = r.Read(b)
		k++
	}
	if err == io.EOF {
		err = nil
	}
	return
}
func ToIndex(k KeyPart, o uint32) Index {
	var i = uint64(binary.BigEndian.Uint32(k[:])) << 32
	i |= uint64(o)
	return Index(i)
}
func FromIndex(i Index) (k KeyPart, o uint32) {
	binary.BigEndian.PutUint32(k[:], uint32(i>>32))
	o = uint32(i & 0xffffffff)
	return
}
func (ids Indexes) Len() int {
	return len(ids)
}

func (ids Indexes) Swap(i, j int) {
	ids[i], ids[j] = ids[j], ids[i]
}

func (ids Indexes) Less(i, j int) bool {
	return ids[i] < ids[j]
}

func (ids Indexes) Find(k KeyPart) (r []uint32) {
	i := ToIndex(k, 0)
	p := ids.FindFirst(i)
	for p < len(ids) {
		ik, o := FromIndex(ids[p])
		if ik != k {
			break
		}
		r = append(r, o)
		p++
	}
	return
}

func (ids Indexes) FindFirst(i Index) int {
	s := 0
	e := len(ids) - 1
	if ids[s] > i {
		return s
	}
	if ids[e] < i {
		return e
	}
	h := (s + e) / 2
	for {
		if ids[h] == i {
			return h
		}
		if ids[h] > i {
			e = h
		}
		if ids[h] < i {
			s = h
		}
		if s == e {
			return s
		}
		if s+1 == e {
			return e
		}
		h = (s + e) / 2
	}
	return -1
}
