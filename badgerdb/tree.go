package badgerdb

import "strings"

var (
	TreePrefix = []byte("t")
)

func keyRawTree(h []byte) []byte {
	return append(TreePrefix, h...)
}

func keyTree(owner, project string) []byte {
	return []byte("t/" + owner + "/" + project)
}

func (self *DB) PutTree(owner, project string, trees []string) error {
	return self.Put(keyTree(owner, project), []byte(strings.Join(trees, "\n")))
}

func (self *DB) PutRawTree(h, b []byte) error {
	return self.Put(keyRawTree(h), b)
}

func (self *DB) GetRawTree(h []byte, cb func([]byte) error) error {
	err := self.Get(keyRawTree(h), cb)
	return err
}

func (self *DB) RawTrees(cb func(k, v []byte) error) error {
	return self.ForEach(TreePrefix, cb)
}

func (self *DB) GetTree(owner, project string) ([]string, error) {
	var ret = []string{}
	err := self.Get(keyTree(owner, project), func(v []byte) error {
		ret = strings.Split(string(v), "\n")
		return nil
	})
	return ret, err
}
