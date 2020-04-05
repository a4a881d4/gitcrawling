package badgerdb

import (
	"io"

	"github.com/a4a881d4/gitcrawling/types"
	"github.com/dgraph-io/badger"
)

var (
	HashPrefix = []byte("hash/")
)

func keyHashRef(h types.Hash) []byte {
	return append(HashPrefix, h[:]...)
}

type HashSession struct {
	txn    *badger.Txn
	it     *badger.Iterator
	db     *DB
	prefix []byte
}

func (db *DB) NewHashSession(sub string) *HashSession {
	txn := db.db.NewTransaction(false)
	opts := badger.DefaultIteratorOptions
	opts.PrefetchSize = 10
	it := txn.NewIterator(opts)
	prefix := append(HashPrefix, []byte(sub)...)
	it.Seek(prefix)
	return &HashSession{
		db:     db,
		it:     it,
		txn:    txn,
		prefix: prefix,
	}
}

func (db *DB) NewHashGeter(sub string) types.SessionedGeter {
	return db.NewHashSession(sub)
}

func (s *HashSession) End() {
	if s.it != nil {
		s.it.Close()
		s.it = nil
	}
	if s.txn != nil {
		s.txn.Discard()
		s.txn = nil
	}
}

func (s *HashSession) Next(prefixlen int, cb func(k, v []byte) error) error {
	if !s.it.ValidForPrefix(s.prefix) {
		return io.EOF
	}
	item := s.it.Item()
	key := item.Key()
	for ; s.it.ValidForPrefix(key[:prefixlen]); s.it.Next() {
		item := s.it.Item()
		var cbk = func(v []byte) error {
			return cb(item.Key(), v)
		}
		err := item.Value(cbk)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *HashSession) NextGroup(prefixlen int, newitem func() types.Byter) (items []types.Byter, err error) {
	var cb = func(k, v []byte) (cberr error) {
		i := newitem()
		cberr = i.SetKey(k)
		if cberr != nil {
			return
		}
		cberr = i.FromByte(v)
		if cberr != nil {
			return
		}
		items = append(items, i)
		return
	}
	err = s.Next(prefixlen, cb)
	return
}
