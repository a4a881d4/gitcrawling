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
	txn *badger.Txn
	it  *badger.Iterator
	db  *DB
}

func (db *DB) NewHashSession() *HashSession {
	txn := db.db.NewTransaction(false)
	opts := badger.DefaultIteratorOptions
	opts.PrefetchSize = 10
	it := txn.NewIterator(opts)
	it.Seek(HashPrefix)
	return &HashSession{
		db:  db,
		it:  it,
		txn: txn,
	}
}

func (s *HashSession) End() {
	if s.txn != nil {
		s.txn.Discard()
		s.txn = nil
	}
	if s.it != nil {
		s.it.Close()
		s.it = nil
	}
}

func (s *HashSession) Next(prefixlen int, cb func(k, v []byte) error) error {
	if !s.it.ValidForPrefix(HashPrefix) {
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

func (s *HashSession) NextGroup(prefixlen int, newitem func() Byter) (items []Byter, err error) {
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
