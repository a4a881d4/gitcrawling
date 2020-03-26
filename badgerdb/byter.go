package badgerdb

type Byter interface {
	ToByte() []byte
	FromByte(v []byte) error
	Key() []byte
	SetKey(k []byte) error
}

func (self *DB) BPut(e Byter) error {
	return self.Put(e.Key(), e.ToByte())
}

func (self *DB) BGet(e Byter) error {
	return self.Get(e.Key(), e.FromByte)
}

func (self *DB) GetRange(prefix []byte, n func() Byter) (es []Byter, err error) {

	var cb = func(k, v []byte) (cberr error) {
		e := n()
		cberr = e.SetKey(k)
		if cberr != nil {
			return
		}
		cberr = e.FromByte(v)
		if cberr != nil {
			return
		}
		es = append(es, e)
		return
	}
	err = self.ForEach(prefix, cb)
	return
}
