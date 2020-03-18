package badgerdb

import (
	badger "github.com/dgraph-io/badger"
)

type DB struct {
	db  *badger.DB
	txn *badger.Txn
}

func NewDB(path string) (*DB, error) {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		db.RunValueLogGC(0.7)
		db, err = badger.Open(badger.DefaultOptions(path))
	}
	return &DB{db, nil}, err
}
func (self *DB) Close() {
	if self.txn != nil {
		self.txn.Commit()
		self.txn.Discard()
	}
	self.db.Close()
}
func (self *DB) NewSession() {
	self.txn = self.db.NewTransaction(true)
}

func (self *DB) Put(k, v []byte) (err error) {
	if err = self.txn.Set(k, v); err == badger.ErrTxnTooBig {
		_ = self.txn.Commit()
		self.txn = self.db.NewTransaction(true)
		err = self.txn.Set(k, v)
	}
	return
}

func (self *DB) EndSession() {
	self.txn.Commit()
	self.txn.Discard()
	self.txn = nil
}

func (self *DB) PutSync(k, v []byte) (err error) {
	err = self.db.Update(func(txn *badger.Txn) error {
		return txn.Set(k, v)
	})
	return
}

func (self *DB) Get(k []byte, cb func(v []byte) error) (err error) {
	err = self.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(k)
		if err != nil {
			return err
		}
		return item.Value(cb)
	})
	return
}
func (self *DB) ForEach(prefix []byte, cb func(k, v []byte) error) error {
	return self.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			var cbk = func(v []byte) error {
				return cb(item.Key(), v)
			}
			err := item.Value(cbk)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
