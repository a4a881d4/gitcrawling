package packext

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"

	"github.com/a4a881d4/gitcrawling/types"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/format/packfile"
)

type HeaderExt [32]byte
type Selector interface {
	Determine([]*ObjEntry) *ObjEntry
}

type maxSelect struct{}
type HashSelect types.Hash

func (maxSelect) Determine(objs []*ObjEntry) (obj *ObjEntry) {
	if len(objs) == 0 {
		return nil
	}
	var size uint32
	for _, v := range objs {
		if v.Size > size {
			obj = v
			size = v.Size
		}
	}
	return
}

func (hash HashSelect) Determine(objs []*ObjEntry) *ObjEntry {
	if len(objs) == 0 {
		return nil
	}
	var maxidx = 0
	for k, v := range objs {
		if v.Hash == plumbing.Hash(hash) && v.Size > objs[maxidx].Size {
			maxidx = k
		}
	}
	return objs[maxidx]
}

var MaxSelect maxSelect

func (h *HeaderExt) Check() bool {
	c := crc32.NewIEEE()
	c.Reset()
	sum := c.Sum(h[:28])
	for k, v := range sum {
		if v != h[28+k] {
			return false
		}
	}
	return true
}

func (h *HeaderExt) ToObjEntry() (*ObjEntry, error) {
	if !h.Check() {
		return nil, fmt.Errorf("Bad Header Ext CRC")
	}
	return NewOEFromBytes(h[:]), nil
}

func (h *HeaderExt) FromFile(r io.Reader) (err error) {
	_, err = io.ReadFull(r, h[:])
	if !h.Check() {
		return fmt.Errorf("Read file headext crc failure")
	}
	return
}

func (h *HeaderExt) ToFile(w io.Writer) (err error) {
	_, err = io.CopyN(w, bytes.NewReader(h[:]), 32)
	return
}

func (h *HeaderExt) Determine(objs []*ObjEntry) *ObjEntry {
	if len(objs) == 0 {
		return nil
	}
	var maxidx = 0
	for k, v := range objs {
		he := v.ToHeaderExt()
		if he == *h && v.Size > objs[maxidx].Size {
			maxidx = k
		}
	}
	return objs[maxidx]
}

type Header struct {
	H *packfile.ObjectHeader
	B *HeaderExt
}

func (h *Header) ToByte() (raw []byte) {
	if h.H.Type.IsDelta() {
		raw = make([]byte, 5+32)
		raw[0] = 0x10
		copy(raw[:5], h.B[:])
	} else {
		raw = make([]byte, 5)
	}
	raw[0] |= (byte(h.H.Type) & 0xf)
	binary.BigEndian.PutUint32(raw[1:], uint32(h.H.Length))
	return
}

func NewHeader(orig, base *ObjEntry) *Header {
	var b HeaderExt
	if orig.OHeader.Type.IsDelta() {
		b = base.ToHeaderExt()
	}
	return &Header{
		H: orig.OHeader,
		B: &b,
	}
}

func HeaderFromByte(raw []byte) (*Header, error) {
	if len(raw) < 5 {
		return nil, fmt.Errorf("too short")
	}
	var b Header

	b.H = &packfile.ObjectHeader{}

	if raw[0]&0x10 != 0 {
		if len(raw) < 5+32 {
			return nil, fmt.Errorf("too short")
		}
		b.H.Type = plumbing.REFDeltaObject
		var base HeaderExt
		copy(base[:], raw[5:5+32])
		if !base.Check() {
			return nil, fmt.Errorf("base not pass crc check")
		}
		b.B = &base
	} else {
		b.H.Type = plumbing.ObjectType(raw[0] & 7)
	}
	var length uint32
	length = binary.BigEndian.Uint32(raw[1:5])
	b.H.Length = int64(length)
	return &b, nil
}

func extFromFile(r io.Reader) (eh *HeaderExt, h *Header, err error) {
	var ehh HeaderExt
	err = ehh.FromFile(r)
	if err != nil {
		return
	}
	eh = &ehh
	var buf = make([]byte, 32+5)
	_, err = r.Read(buf[:1])
	if err != nil {
		return
	}
	if buf[0]&0x10 != 0 {
		_, err = io.ReadFull(r, buf[1:])
	} else {
		_, err = io.ReadFull(r, buf[1:5])
	}
	if err != nil {
		return
	}
	h, err = HeaderFromByte(buf)
	if err != nil {
		return
	}
	return
}

func ExtFromFile(r io.Reader) (obj *ObjEntry, readed int, err error) {
	var peh *HeaderExt
	var ph *Header
	peh, ph, err = extFromFile(r)
	if err != nil {
		return
	}
	obj = NewOEFromBytes((*peh)[:])
	obj.OHeader = ph.H
	if ph.H.Type.IsDelta() {
		readed = 32 + 5 + 32
	} else {
		readed = 32 + 5
	}
	return
}
