package packext

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sync"

	"github.com/a4a881d4/gitcrawling/types"
	"gopkg.in/src-d/go-git.v4/utils/ioutil"
)

type Packns struct {
	Servered []types.Classifer
	Ids      []types.Indexes
	Pfs      []*os.File
	Dir      string
}

func NewPackNo(s []byte, d string) (r *Packns, err error) {
	c := make([]types.Classifer, len(s))
	for k, v := range s {
		c[k] = ByteClassify(v)
	}
	r = &Packns{
		Servered: c,
		Dir:      d,
		Ids:      make([]types.Indexes, len(s)),
		Pfs:      make([]*os.File, len(s)),
	}
	for k, v := range r.Servered {
		r.Pfs[k], err = os.Open(path.Join(r.Dir, v.FileNamePrefix()+"no.pack"))
		if err != nil {
			return
		}
		var idxr *os.File
		idxr, err = os.Open(path.Join(r.Dir, v.FileNamePrefix()+"no.idx"))
		if err != nil {
			return
		}
		defer idxr.Close()
		r.Ids[k], err = types.IndexFromFile(idxr)
		if err != nil {
			return
		}
	}
	return
}
func (r *Packns) Close() (err error) {
	for _, v := range r.Pfs {
		err = v.Close()
		if err != nil {
			return
		}
	}
	return
}

var (
	ErrorNoFind = errors.New("No Find")
)

type HeadOffset struct {
	Head   [32]byte
	Offset uint32
}

func (r *Packns) GetByHash(h types.Hash) (head map[int][]HeadOffset, err error) {
	head = make(map[int][]HeadOffset)
	var key types.KeyPart
	for k, v := range r.Servered {
		if v.Hit(h) {
			copy(key[:], h[1:5])
			p := r.Ids[k].Find(key)
			for _, o := range p {
				var h HeadOffset
				h.Offset = o
				_, err = r.Pfs[k].Seek(int64(o), 0)
				if err != nil {
					return
				}
				_, err = r.Pfs[k].Read(h.Head[:])
				if err != nil {
					return
				}
				head[k] = append(head[k], h)
			}
		}
	}
	return
}
func (r *Packns) GetAnyHash(h types.Hash) (body []byte, d bool, base [32]byte, err error) {
	var head map[int][]HeadOffset
	head, err = r.GetByHash(h)
	var mh = make(map[uint32]*HeadOffset)
	for _, v := range head {
		for _, hh := range v {
			o := NewOEFromBytes(hh.Head[:])
			mh[o.Size] = &hh
		}
	}
	var max uint32
	for k, _ := range mh {
		if k > max {
			max = k
		}
	}
	return r.GetHashAndLen(mh[max].Head, head)
}
func (r *Packns) GetSpecial(h [32]byte) (body []byte, d bool, base [32]byte, err error) {
	var head map[int][]HeadOffset
	var find bool
	find = false
	o := NewOEFromBytes(h[:])
	head, err = r.GetByHash(types.Hash(o.Hash))
	for _, v := range head {
		for _, hh := range v {
			if hh.Head == h {
				find = true
			}
		}
	}
	if find {
		return r.GetHashAndLen(h, head)
	} else {
		err = ErrorNoFind
		return
	}
}

func (r *Packns) GetHashAndLen(h [32]byte, head map[int][]HeadOffset) (body []byte, deta bool, base [32]byte, err error) {
	var prefact *HeadOffset
	var fk = -1
	for k, v := range head {
		for _, hh := range v {
			if hh.Head == h {
				fk = k
				prefact = &hh
			}
		}
	}
	if fk == -1 {
		err = ErrorNoFind
		return
	}
	r.Pfs[fk].Seek(int64(prefact.Offset+32), 0)
	var ihead [5]byte
	var n int
	n, err = r.Pfs[fk].Read(ihead[:])
	if err != nil {
		return
	}
	if n != 5 {
		err = fmt.Errorf("Read fail %d", n)
		return
	}
	dsize := binary.BigEndian.Uint32(ihead[1:5])
	if ihead[0]&0x10 != 0 {
		deta = true
		n, err = r.Pfs[fk].Read(base[:])
		if err != nil {
			return
		}
		if n != 32 {
			err = fmt.Errorf("Read fail %d", n)
			return
		}
	} else {
		deta = false
	}

	bufw := new(bytes.Buffer)
	bufw.Reset()
	zr := zlibReaderPool.Get().(io.ReadCloser)
	defer zlibReaderPool.Put(zr)

	if err = zr.(zlib.Resetter).Reset(r.Pfs[fk], nil); err != nil {
		err = fmt.Errorf("zlib reset error: %s", err)
		return
	}

	defer ioutil.CheckClose(zr, &err)
	buf := byteSlicePool.Get().([]byte)
	_, err = io.CopyBuffer(bufw, zr, buf)
	byteSlicePool.Put(buf)

	body = bufw.Bytes()
	if len(body) != int(dsize) {
		err = fmt.Errorf("Wrony size, want %d, but %d", dsize, len(body))
	}
	return
}

var zlibInitBytes = []byte{0x78, 0x9c, 0x01, 0x00, 0x00, 0xff, 0xff, 0x00, 0x00, 0x00, 0x01}

var zlibReaderPool = sync.Pool{
	New: func() interface{} {
		r, _ := zlib.NewReader(bytes.NewReader(zlibInitBytes))
		return r
	},
}
var byteSlicePool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 32*1024)
	},
}
