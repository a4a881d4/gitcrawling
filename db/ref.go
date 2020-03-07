package db
import (
	"github.com/a4a881d4/gitcrawling/gitext"
	"github.com/ethereum/go-ethereum/ethdb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type RefDB struct {
	db *leveldb.DB
}

func NewRefDB(db *DB) *RefDB {
	return &RefDB{db.GetDB()}
}
func(self *RefDB) Close() {
	self.db.Close()
}
func keyRef(owner, project string) []byte {
	return []byte("r/" + owner + "/" + project)
}

func(self *RefDB) PutRefSync(owner, project string, r []gitext.Ref) (err error) {
	opts := &opt.WriteOptions{Sync: true}
	err = self.db.Put(keyRef(owner, project), gitext.NewRefRecode(r).DecodeRef(), opts)
	return
}

func(self *RefDB) HasRef(owner, project string) (bool, error) {
	return self.db.Has(keyRef(owner, project),nil)
}

func(self *RefDB) DelRef(owner, project string) error {
	return self.db.Delete(keyRef(owner, project),nil)
}

func (self *RefDB) GetRef(owner, project string) []gitext.Ref {
	buf, err := self.db.Get(keyRef(owner, project), nil)
	if err != nil {
		return []gitext.Ref{}
	}
	record := gitext.EncodeRef(buf)
	return record.Refs
}

func (self *RefDB) getRef(owner, project string) (*gitext.RefRecode, error) {
	buf, err := self.db.Get(keyRef(owner, project), nil)
	if err != nil {
		return nil, err
	}
	record := gitext.EncodeRef(buf)
	return record, err
}

func(self *RefDB) OK(owner, project string) bool {
	r,err := getRef(owner, project)
	if err!=nil {
		return false
	} else {
		return r.OK()
	}
}

func(self *RefDB) RemoteOK(owner, project string) bool {
	r,err := getRef(owner, project)
	if err!=nil {
		return true
	} else {
		return r.RemoteOK()
	}
}
