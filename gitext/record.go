package gitext

import (
	"encoding/json"
)

type Ref struct {
	Name string
	Hash string
}
type Record struct {
	Language string
	Star     uint64
	Refs     []Ref
	Last     uint64
}

func EmptyRef() []Ref {
	return []Ref{}
}

func (r *Record) String() string {
	if buf, err := json.MarshalIndent(r, "", "  "); err != nil {
		return "Bad Record"
	} else {
		return string(buf)
	}
}
