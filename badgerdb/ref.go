package badgerdb

var (
	RefPrefix = []byte("r/")
)

func keyRawRef(h []byte) []byte {
	return append(RefPrefix, h...)
}

func (self *DB) PutRawRef(h, b []byte) error {
	return self.Put(keyRawRef(h), b)
}

func (self *DB) GetRawRef(h []byte, cb func([]byte) error) error {
	err := self.Get(keyRawRef(h), cb)
	return err
}

func (self *DB) RawRefs(cb func(k, v []byte) error) error {
	return self.ForEach(RefPrefix, cb)
}
