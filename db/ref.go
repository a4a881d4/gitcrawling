package db

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/a4a881d4/gitcrawling/gitext"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type RefDB struct {
	Path      string
	db        *leveldb.DB
	Retry     int
	cache     map[string]*gitext.RefRecord
	dirty     []string
	withCache bool
}

func NewRefDB(path string) *RefDB {
	return &RefDB{
		Path:      path,
		Retry:     10,
		cache:     make(map[string]*gitext.RefRecord),
		dirty:     make([]string, 0),
		withCache: true,
	}
}

func (self *RefDB) NoDB() {
	self.withCache = true
}
func (self *RefDB) open(retry int) *leveldb.DB {
	if retry >= self.Retry {
		self.db = nil
		return self.db
	}
	opts := &opt.Options{}
	db, err := leveldb.OpenFile(self.Path, opts)
	if err != nil {
		self.db = nil
		<-time.After(time.Duration(rand.Intn(1<<(8+retry))+(1<<(8+retry))) * time.Millisecond)
		fmt.Println("Retry Open", retry, err)
		return self.open(retry + 1)
	}
	self.db = db
	return db
}

func (self *RefDB) Open() *leveldb.DB {
	return self.open(0)
}

func (self *RefDB) Close() {
	defer func() {
		self.db = nil
	}()
	if self.db != nil {
		self.db.Close()
	}
}

func skeyRef(owner, project string) string {
	return "r/" + owner + "/" + project
}
func keyRef(owner, project string) []byte {
	return []byte(skeyRef(owner, project))
}

func (self *RefDB) PutRefSync(owner, project string, r []gitext.Ref) (err error) {
	defer self.Close()
	opts := &opt.WriteOptions{Sync: true}
	err = self.Open().Put(keyRef(owner, project), gitext.NewRefRecord(r).DecodeRef(), opts)
	return
}

func (self *RefDB) Stop() {
	self.flush()
}

func (self *RefDB) flush() {
	if len(self.dirty) > 0 {
		defer self.Close()
		opts := &opt.WriteOptions{Sync: true}
		fmt.Println("flush open db")
		fmt.Println(self.dirty, len(self.dirty))
		self.Open()
		fmt.Println("flush open done")
		for _, k := range self.dirty {
			self.db.Put([]byte(k), self.cache[k].DecodeRef(), opts)
		}
		self.dirty = make([]string, 0)
	}
}

func (self *RefDB) PutRef(owner, project string, r []gitext.Ref) (err error) {
	k := skeyRef(owner, project)
	self.cache[k] = gitext.NewRefRecord(r)
	self.dirty = append(self.dirty, k)
	if len(self.dirty) > 20 {
		self.flush()
	}
	return nil
}

func (self *RefDB) SetBuild(owner, project string, ir []gitext.Ref) (r []gitext.Ref, err error) {
	defer self.Close()
	if len(ir) == 0 {
		ir = self.GetRef(owner, project)
	}
	k := skeyRef(owner, project)
	self.cache[k] = gitext.NewBuildRecord(r)
	self.dirty = append(self.dirty, k)
	if len(self.dirty) > 20 {
		self.flush()
	}
	r = ir
	return
}

func (self *RefDB) HasRef(owner, project string) (bool, error) {
	k := skeyRef(owner, project)
	if _, ok := self.cache[k]; ok {
		return ok, nil
	}
	if self.withCache {
		return false, fmt.Errorf("Not In Cache")
	}
	if _, err := self.getRef(owner, project); err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func (self *RefDB) DelRef(owner, project string) error {
	k := skeyRef(owner, project)
	if _, ok := self.cache[k]; ok {
		delete(self.cache, k)
	}
	defer self.Close()
	return self.Open().Delete(keyRef(owner, project), nil)
}

func (self *RefDB) GetRef(owner, project string) []gitext.Ref {
	k := skeyRef(owner, project)
	if r, ok := self.cache[k]; ok {
		return r.Refs
	}
	if self.withCache {
		return []gitext.Ref{}
	}
	record, err := self.getRef(owner, project)
	if err != nil {
		return []gitext.Ref{}
	}
	return record.Refs
}

func (self *RefDB) getRef(owner, project string) (*gitext.RefRecord, error) {
	defer self.Close()
	k := skeyRef(owner, project)
	buf, err := self.Open().Get(keyRef(owner, project), nil)
	if err != nil {
		return nil, err
	}
	record := gitext.EncodeRef(buf)
	if err != nil {
		self.cache[k] = record
	}
	return record, err
}

func (self *RefDB) OK(owner, project string) bool {
	k := skeyRef(owner, project)
	if r, ok := self.cache[k]; ok {
		return r.OK()
	}
	if self.withCache {
		return false
	}
	r, err := self.getRef(owner, project)
	if err != nil {
		return false
	} else {
		return r.OK()
	}
}

func (self *RefDB) RemoteOK(owner, project string) bool {
	k := skeyRef(owner, project)
	if r, ok := self.cache[k]; ok {
		return r.RemoteOK()
	}
	if self.withCache {
		return true
	}
	r, err := self.getRef(owner, project)
	if err != nil {
		return true
	} else {
		return r.RemoteOK()
	}
}

func (self *RefDB) IsBuild(owner, project string) bool {
	k := skeyRef(owner, project)
	if r, ok := self.cache[k]; ok {
		return r.IsBuild()
	}
	if self.withCache {
		return false
	}
	r, err := self.getRef(owner, project)
	if err != nil {
		return false
	} else {
		return r.IsBuild()
	}
}

func (self *RefDB) CashePrefech(r []string) {
	self.flush()
	self.Open()
	defer self.Close()
	for _, v := range r {
		k := "r/" + v
		buf, err := self.db.Get([]byte(k), nil)
		if err == nil {
			record := gitext.EncodeRef(buf)
			self.cache[k] = record
		}
	}
}
func (self *RefDB) UnCashe(r []string) {
	for _, v := range r {
		k := "r/" + v
		delete(self.cache, k)
	}
}

func (self *RefDB) Init(r []string) {
	self.flush()
	self.Open()
	defer self.Close()
	self.db.CompactRange(util.Range{nil, nil})

	self.cache = make(map[string]*gitext.RefRecord)
	for _, v := range r {
		k := "r/" + v
		buf, err := self.db.Get([]byte(k), nil)
		if err == nil {
			record := gitext.EncodeRef(buf)
			self.cache[k] = record
		}
	}
}
