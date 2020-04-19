package topic

import (
	"errors"
	"io"
	"os"
	"path"
	"sync"

	"github.com/a4a881d4/gitcrawling/packext"
	"github.com/a4a881d4/gitcrawling/types"
)

type ServerChannel struct {
	Lock  sync.Mutex
	Class types.Classifer
	Ids   types.Indexes
	Pfs   *os.File
	Dir   string
	Sel   packext.Selector
}

var (
	ErrorNoFind = errors.New("No Find")
)

func NewServerByteChannels(dir string) (m []*ServerChannel, err error) {
	for i := 0; i < 256; i++ {
		var c *ServerChannel
		c, err = NewServerChannel(dir, packext.ByteClassify(i))
		if err != nil {
			return
		}
		m = append(m, c)
	}
	return
}

func NewServerChannel(dir string, c types.Classifer) (m *ServerChannel, err error) {
	m = &ServerChannel{
		Dir:   dir,
		Class: c,
		Sel:   packext.MaxSelect,
	}
	fnb := m.Class.NamePrefix()
	fn := path.Join(m.Dir, fnb+"no.pack")
	if _, err = os.Stat(fn); err != nil {
		m.Pfs, err = os.Create(fn)
		if err != nil {
			return nil, err
		}
	} else {
		m.Pfs, err = os.OpenFile(fn, os.O_APPEND|os.O_RDWR, 0644)
		if err != nil {
			return nil, err
		}
	}
	var idxr *os.File
	idxr, err = os.Open(path.Join(m.Dir, fnb+"no.idx"))
	if err != nil {
		return
	}
	defer idxr.Close()
	m.Ids, err = types.IndexFromFile(idxr)
	if err != nil {
		return
	}
	return m, nil
}

func (m *ServerChannel) Close() {
	m.Pfs.Close()
}

func (m *ServerChannel) getObj(hash types.Hash, sel packext.Selector) (obj *packext.ObjEntry, err error) {
	var key types.KeyPart
	var objs []*packext.ObjEntry
	var Head [32]byte
	if m.Class.Hit(hash) {
		copy(key[:], hash[1:5])
		p := m.Ids.Find(key)
		for _, o := range p {
			_, err = m.Pfs.Seek(int64(o), 0)
			if err != nil {
				return
			}
			_, err = io.ReadFull(m.Pfs, Head[:])
			if err != nil {
				return
			}
			obj = packext.NewOEFromBytes(Head[:])
			obj.Offset = uint64(o)
			if types.Hash(obj.Hash) == hash {
				objs = append(objs, obj)
			}
		}
	}
	if len(objs) == 0 {
		err = ErrorNoFind
		return
	}
	obj = sel.Determine(objs)
	return
}

func (m *ServerChannel) GetSepcial(h *packext.HeaderExt) (raw []byte, err error) {
	var o, fo *packext.ObjEntry
	o, err = h.ToObjEntry()
	fo, err = m.getObj(types.Hash(o.Hash), h)
	raw = make([]byte, int(fo.Size))
	_, err = m.Pfs.Seek(int64(o.Offset), 0)
	if err != nil {
		return
	}
	_, err = io.ReadFull(m.Pfs, raw)
	return
}

func (m *ServerChannel) GetByHash(h types.Hash) (raw []byte, err error) {
	var o, fo *packext.ObjEntry
	fo, err = m.getObj(h, packext.MaxSelect)
	raw = make([]byte, int(fo.Size))
	_, err = m.Pfs.Seek(int64(o.Offset), 0)
	if err != nil {
		return
	}
	_, err = io.ReadFull(m.Pfs, raw)
	return
}
