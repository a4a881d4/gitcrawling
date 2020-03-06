package db

import (
	"strings"

	"github.com/a4a881d4/gitcrawling/gitext"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type DB struct {
	db *leveldb.DB
}

func NewDB(dir string) (*DB, error) {
	opts := &opt.Options{OpenFilesCacheCapacity: 5}
	db, err := leveldb.OpenFile(dir, opts)
	return &DB{db}, err
}

func keyRepo(owner, project string) []byte {
	return []byte("R/" + owner + "/" + project)
}

func (self *DB) GetRepo(owner, project string) (*gitext.Record, error) {
	var r gitext.Record
	rlpRecord, err := self.db.Get(keyRepo(owner, project), nil)
	if err != nil {
		return nil, err
	}
	err = rlp.DecodeBytes(rlpRecord, &r)
	return &r, err
}

func (self *DB) PutRepo(owner, project string, r *gitext.Record) (err error) {
	var buf []byte
	if buf, err = rlp.EncodeToBytes(r); err != nil {
		return
	}
	err = self.db.Put(keyRepo(owner, project), buf, nil)
	return
}
func (self *DB) PutRepoSync(owner, project string, r *gitext.Record) (err error) {
	var buf []byte
	if buf, err = rlp.EncodeToBytes(r); err != nil {
		return
	}
	opts := &opt.WriteOptions{Sync: true}
	err = self.db.Put(keyRepo(owner, project), buf, opts)
	return
}
func (self *DB) Close() {
	self.db.Close()
}

func (self *DB) ForEach(prefix string, cb func(k, v []byte)) error {
	iter := self.db.NewIterator(util.BytesPrefix([]byte(prefix)), nil)
	defer iter.Release()
	for iter.Next() {
		cb(iter.Key(), iter.Value())
	}
	return iter.Error()
}

func (self *DB) ForEachRepo(cb func(owner, project string, r *gitext.Record)) (err error) {
	self.ForEach("R/", func(k, v []byte) {
		ks := strings.Split(string(k), "/")
		var r gitext.Record
		err = rlp.DecodeBytes(v, &r)
		cb(ks[1], ks[2], &r)
	})
	return nil
}

func keyRef(owner, project string) []byte {
	return []byte("r/" + owner + "/" + project)
}

func(self *DB) PutRefSync(owner, project string, r []gitext.Ref) (err error) {
	var buf []byte
	if buf, err = rlp.EncodeToBytes(r); err != nil {
		return
	}
	opts := &opt.WriteOptions{Sync: true}
	err = self.db.Put(keyRef(owner, project), buf, opts)
	return
}

func(self *DB) HasRef(owner, project string) (bool, error) {
	return self.db.Has(keyRef(owner, project),nil)
}

func(self *DB) DelRef(owner, project string) error {
	return self.db.Delete(keyRef(owner, project),nil)
}

func (self *DB) GetRef(owner, project string) ([]gitext.Ref, error) {
	var r []gitext.Ref
	rlpRecord, err := self.db.Get(keyRef(owner, project), nil)
	if err != nil {
		return nil, err
	}
	err = rlp.DecodeBytes(rlpRecord, &r)
	return r, err
}
