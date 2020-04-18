package badgerdb

import (
	"fmt"
	"io"
	"strings"

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
func (tdb *DB) Group(idx, pos int) (map[string][]string, error) {
	r := make(map[string][]string)
	fmt.Println("Begin Group", idx, pos)
	fmt.Printf("Progress:\033[s")
	counter := 0
	err := tdb.ForEach([]byte("hash/"), func(k, v []byte) error {
		ss := strings.Split(string(k), "/")
		if len(ss) > idx && len(ss) > pos {
			r[ss[idx]] = append(r[ss[idx]], ss[pos])
		}
		counter++
		if counter%1000 == 0 {
			fmt.Printf("\033[u\033[K%20d", counter)
		}
		return nil
	})
	fmt.Println("End Group")
	return r, err
}
