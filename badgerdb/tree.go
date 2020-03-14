package badgerdb

import "strings"

func keyTree(owner, project string) []byte {
	return []byte("t/" + owner + "/" + project)
}

func (self *DB) PutTree(owner, project string, trees []string) error {
	return self.Put(keyTree(owner, project), []byte(strings.Join(trees, "\n")))
}

func (self *DB) GetTree(owner, project string) ([]string, error) {
	var ret = []string{}
	err := self.Get(keyTree(owner, project), func(v []byte) error {
		ret = strings.Split(string(v), "\n")
		return nil
	})
	return ret, err
}
