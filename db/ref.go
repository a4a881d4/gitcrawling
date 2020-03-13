package db
import (
	"fmt"
	"time"

	"github.com/a4a881d4/gitcrawling/gitext"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type RefDB struct {
	Path string
	db   *leveldb.DB
	Retry int
	cache map[string]*gitext.RefRecord
	dirty []string
}

func NewRefDB(path string) *RefDB {
	return &RefDB{
		Path:path,
		Retry:10,
		cache : make(map[string]*gitext.RefRecord),
		dirty : make([]string,20),
	}
}

func(self *RefDB) open(retry int) *leveldb.DB {
	if retry>=self.Retry {
		self.db = nil
		return self.db
	}
	opts := &opt.Options{OpenFilesCacheCapacity: 5}
	db, err := leveldb.OpenFile(self.Path, opts)
	if err != nil {
		self.db = nil
		<- time.After(time.Duration(1<<(8+retry)) * time.Millisecond)
		fmt.Println("Retry Open",retry)
		return self.open(retry+1)
	}
	self.db = db
	return db
}

func(self *RefDB) Open() *leveldb.DB {
	return self.open(0)
}

func(self *RefDB) Close() {
	defer func(){
		self.db = nil
	}()
	if self.db!=nil {
		self.db.Close()
	}
}

func skeyRef(owner, project string) string {
	return "r/" + owner + "/" + project
}
func keyRef(owner, project string) []byte {
	return []byte(skeyRef(owner, project))
}

func(self *RefDB) PutRefSync(owner, project string, r []gitext.Ref) (err error) {
	defer self.Close()
	opts := &opt.WriteOptions{Sync: true}
	err = self.Open().Put(keyRef(owner, project), gitext.NewRefRecord(r).DecodeRef(), opts)
	return
}

func(self *RefDB) Stop() {
	self.flush()
}

func(self *RefDB) flush() {
	defer self.Close()
	opts := &opt.WriteOptions{Sync: true}
	self.Open()
	for _,k := range self.dirty {
		self.db.Put([]byte(k),self.cache[k].DecodeRef(),opts)
	}
	self.dirty = make([]string,20)
}

func(self *RefDB) PutRef(owner, project string, r []gitext.Ref) (err error) {
	k := skeyRef(owner, project)
	self.cache[k] = gitext.NewRefRecord(r)
	self.dirty = append(self.dirty,k)
	if len(self.dirty)>20 {
		self.flush()
	}
	return nil
}

func(self *RefDB) SetBuild(owner, project string, ir []gitext.Ref) (r []gitext.Ref,err error) {
	defer self.Close()
	if len(ir)==0 {
		ir = self.GetRef(owner, project)
	}
	opts := &opt.WriteOptions{Sync: true}
	err = self.Open().Put(keyRef(owner, project), gitext.NewBuildRecord(r).DecodeRef(), opts)
	r = ir
	return
}

func(self *RefDB) HasRef(owner, project string) (bool, error) {
	k := skeyRef(owner, project)
	if _,ok := self.cache[k];ok {
		return ok,nil
	}
	if _,err := self.getRef(owner, project); err != nil {
		return false,err
	} else {
		return true,nil
	}
}

func(self *RefDB) DelRef(owner, project string) error {
	k := skeyRef(owner, project)
	if _,ok := self.cache[k];ok {
		delete(self.cache,k)
	}
	defer self.Close()
	return self.Open().Delete(keyRef(owner, project),nil)
}

func (self *RefDB) GetRef(owner, project string) []gitext.Ref {
	k := skeyRef(owner, project)
	if r,ok := self.cache[k];ok {
		return r.Refs
	}
	record,err := self.getRef(owner, project)
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
		self.cache[k]=record
	}
	return record, err
}

func(self *RefDB) OK(owner, project string) bool {
	k := skeyRef(owner, project)
	if r,ok := self.cache[k];ok {
		return r.OK()
	}
	r,err := self.getRef(owner, project)
	if err!=nil {
		return false
	} else {
		return r.OK()
	}
}

func(self *RefDB) RemoteOK(owner, project string) bool {
	k := skeyRef(owner, project)
	if r,ok := self.cache[k];ok {
		return r.RemoteOK()
	}
	r,err := self.getRef(owner, project)
	if err!=nil {
		return true
	} else {
		return r.RemoteOK()
	}
}

func(self *RefDB) IsBuild(owner, project string) bool {
	k := skeyRef(owner, project)
	if r,ok := self.cache[k];ok {
		return r.IsBuild()
	}
	r,err := self.getRef(owner, project)
	if err!=nil {
		return false
	} else {
		return r.IsBuild()
	}
}

func(self *RefDB) Init(r []string) {
	defer self.Close()
	self.Open()
	for _,v := range r {
		k := "r/" +v
		buf, err := self.db.Get([]byte(k), nil)
		if err == nil {
			record := gitext.EncodeRef(buf)
			self.cache[k]=record
		}
	}
}