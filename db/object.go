package db

import (
	"bytes"
	"fmt"
	"io"

	"github.com/syndtr/goleveldb/leveldb"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	gitutil "gopkg.in/src-d/go-git.v4/utils/ioutil"
)

type ObjDB struct {
	db *leveldb.DB
	msg chan string
	Force bool
}

func NewObjectDB(db *DB) *ObjDB {
	msg := make(chan string,10)
	return &ObjDB{db.GetDB(),msg,false}
}

func(self *ObjDB) Close() {
	self.db.Close()
}

func keyBlob(hash plumbing.Hash) []byte {
	return append([]byte("b/"),hash[:]...)
}

func (self *ObjDB) PutBlob(hash plumbing.Hash, v []byte) (err error) {
	err = self.db.Put(keyBlob(hash), v, nil)
	return
}

func(self *ObjDB) GetBlobByHex(h string) ([]byte,error) {
	return self.db.Get(keyBlob(plumbing.NewHash(h)),nil)
}
func(self *ObjDB) HasBlob(hash plumbing.Hash) (bool, error) {
	return self.db.Has(keyBlob(hash),nil)
}

func(self *ObjDB) PutObj(b *object.Blob) error {
	k := b.ID()
	if !self.Force {
		has,err := self.HasBlob(k)
		if  has || err!=nil {
			if has {
				self.msg <- fmt.Sprintf("W: has obj %s %v",k.String(),err)
			}
			return nil
		}
	}
	
	r,err := b.Reader()
	if err != nil {
		self.msg <- fmt.Sprintf("E: get reader failure %s",k.String())
		return nil
	}
	defer gitutil.CheckClose(r,&err)

	buf   := bytes.NewBuffer(make([]byte,b.Size))
	s,err := io.Copy(buf,r)
	if int64(s)!=b.Size {
		self.msg <- fmt.Sprintf("E: Wrony size %s %d:%d",k.String(),s,b.Size)
		return nil
	}

	if err != nil {
		self.msg <- fmt.Sprintf("E: io copy failure %s",k.String())
		return nil
	}
	if err = self.PutBlob(k,buf.Bytes()); err !=nil {
		self.msg <- fmt.Sprintf("E: put to db failure %s",k.String())
		return nil
	}
	self.msg <- fmt.Sprintf("I: done %s",k.String())
	return nil
}

func(self *ObjDB) PutObjects(iter *object.BlobIter) error {
	err := iter.ForEach(self.PutObj)
	self.msg <- fmt.Sprintf("I: Finish")
	return err
}

func(self *ObjDB) Msg() chan string {
	return self.msg
}

func(self *ObjDB) Wait(v bool,show int) {
	var c int
	c = 0
	for{
		info :=<- self.msg
		switch{
		case info[:2]=="E:":
			fmt.Println("")
			fmt.Println("Error:",info[2:])
		case info[:2]=="W:":
			if v{
				fmt.Println("")
				fmt.Println("Warning:",info[2:])
			}
		case info[:2]=="F:":
			fmt.Println("")
			fmt.Println("Warning:",info[2:])
			return
		case info[:2]=="I:":
			c++
			if c%show==0 {
				fmt.Printf("*")
			}
			return
		}
	}
}