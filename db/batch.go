package db

import (
	"github.com/a4a881d4/gitcrawling/types"
	"github.com/syndtr/goleveldb/leveldb"
)

type BatchDB struct {
	db    *leveldb.DB
	batch *leveldb.Batch
}

func NewBatchDB(db *DB) *BatchDB {
	return &BatchDB{db.GetDB(), nil}
}

func (self *BatchDB) Close() {
	self.EndSession()
	self.db.Close()
}

func (self *BatchDB) NewSession() error {
	if self.batch != nil {
		err := self.db.Write(self.batch, nil)
		self.batch = nil
		if err != nil {
			return err
		}
	}
	self.batch = new(leveldb.Batch)
	return nil
}

func (self *BatchDB) EndSession() error {
	if self.batch != nil {
		err := self.db.Write(self.batch, nil)
		self.batch = nil
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *BatchDB) Put(k, v []byte) {
	self.batch.Put(k, v)
}

func (self *BatchDB) BPut(e types.Byter) {
	self.Put(e.Key(), e.ToByte())
}
