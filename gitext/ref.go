package gitext

import (
	"encoding/json"
	"time"
)

type Ref struct {
	Name string
	Hash string
}

type RefRecord struct {
	LocalOK  bool
	ServerOK bool
	Build    bool
	Refs     []Ref
	Last     time.Time
}

func(r *RefRecord) DecodeRef() []byte {
	buf,_ := json.Marshal(r)
	return buf
}

func EncodeRef(buf []byte) *RefRecord {
	var r = RefRecord{}
	json.Unmarshal(buf,&r)
	return &r
}

func(r *RefRecord) OK() bool {
	return r.LocalOK
}

func(r *RefRecord) RemoteOK() bool {
	return r.ServerOK
}

func(r *RefRecord) LastSeen() time.Time {
	return r.Last
}

func(r *RefRecord) IsBuild() bool {
	return r.Build
}

func NewBuildRecord(refs []Ref) *RefRecord {
	r := NewRefRecord(refs)
	r.Build = true
	return r
}

func NewRefRecord(refs []Ref) *RefRecord {
	remoteOk := len(refs)>0
	return &RefRecord{
		Last    : time.Now(),
		LocalOK : true,
		ServerOK: remoteOk,
		Build   : false,
		Refs    : refs,
	}
}
