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
}

func NewRefDB(path string) *RefDB {
	return &RefDB{
		Path:path,
		Retry:10,
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
		<- time.After(time.Second*1)
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

func keyRef(owner, project string) []byte {
	return []byte("r/" + owner + "/" + project)
}

func(self *RefDB) PutRefSync(owner, project string, r []gitext.Ref) (err error) {
	defer self.Close()
	opts := &opt.WriteOptions{Sync: true}
	err = self.Open().Put(keyRef(owner, project), gitext.NewRefRecord(r).DecodeRef(), opts)
	return
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
	defer self.Close()
	return self.Open().Has(keyRef(owner, project),nil)
}

func(self *RefDB) DelRef(owner, project string) error {
	defer self.Close()
	return self.Open().Delete(keyRef(owner, project),nil)
}

func (self *RefDB) GetRef(owner, project string) []gitext.Ref {
	record,err := self.getRef(owner, project)
	if err != nil {
		return []gitext.Ref{}
	}
	return record.Refs
}

func (self *RefDB) getRef(owner, project string) (*gitext.RefRecord, error) {
	defer self.Close()
	buf, err := self.Open().Get(keyRef(owner, project), nil)
	if err != nil {
		return nil, err
	}
	record := gitext.EncodeRef(buf)
	return record, err
}

func(self *RefDB) OK(owner, project string) bool {
	r,err := self.getRef(owner, project)
	if err!=nil {
		return false
	} else {
		return r.OK()
	}
}

func(self *RefDB) RemoteOK(owner, project string) bool {
	r,err := self.getRef(owner, project)
	if err!=nil {
		return true
	} else {
		return r.RemoteOK()
	}
}

func(self *RefDB) IsBuild(owner, project string) bool {
	r,err := self.getRef(owner, project)
	if err!=nil {
		return false
	} else {
		return r.IsBuild()
	}
}