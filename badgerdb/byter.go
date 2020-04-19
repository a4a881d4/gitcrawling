package badgerdb

import "github.com/a4a881d4/gitcrawling/types"

func (self *DB) BPut(e types.Byter) error {
	return self.Put(e.Key(), e.ToByte())
}

func (self *DB) BGet(e types.Byter) error {
	return self.Get(e.Key(), e.FromByte)
}

func (self *DB) GetRange(prefix []byte, n func() types.Byter) (es []types.Byter, err error) {

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
