package packext

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/a4a881d4/gitcrawling/types"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/format/idxfile"
	"gopkg.in/src-d/go-git.v4/plumbing/format/packfile"
)

var (
	ErrSize    = errors.New("Size Error")
	ErrMiss    = errors.New("Miss")
	ErrKey     = errors.New("Wrony Key Format")
	ErrBadName = errors.New("Bad Pack File Name")
)

type ObjEntry struct {
	idxfile.Entry
	Size     uint32
	PackFile OriginPackFile
	OHeader  *packfile.ObjectHeader
	RealType plumbing.ObjectType
}

type OriginPackFile types.Hash

type OriginPackFiles map[OriginPackFile]string

func NewOriginPackFiles() OriginPackFiles {
	return make(map[OriginPackFile]string)
}

var DefaultOPS = NewOriginPackFiles()

func (ops OriginPackFiles) GetHash(fn string) (OriginPackFile, error) {
	sfn := strings.Replace(fn, `\`, `/`, -1)
	baseName := path.Base(sfn)
	fmt.Println("BaseName", baseName)
	if baseName[:5] != "pack-" || len(baseName) < 45 {
		return OriginPackFile(plumbing.ZeroHash), ErrBadName
	}
	h, err := hex.DecodeString(baseName[5:45])
	if err != nil {
		return OriginPackFile(plumbing.ZeroHash), err
	}
	var hash OriginPackFile
	copy(hash[:], h[:])
	if _, ok := ops[hash]; !ok {
		ops[hash] = fn
	}
	return hash, nil
}
func (ops OriginPackFiles) GetFileName(hash OriginPackFile) (string, error) {
	if fn, ok := ops[hash]; ok {
		return fn, nil
	} else {
		return "", ErrMiss
	}
}
func (o OriginPackFile) NewEmptyEntry() *ObjEntry {
	return &ObjEntry{PackFile: o}
}

func (o OriginPackFile) NewEntry(e *idxfile.Entry) *ObjEntry {
	return &ObjEntry{
		Entry:    *e,
		PackFile: o,
	}
}
func (o OriginPackFile) String() string {
	return plumbing.Hash(o).String()
}
func (obj *ObjEntry) ToByte() []byte {
	buf := make([]byte, 16)
	binary.BigEndian.PutUint64(buf[:8], obj.Offset)
	binary.BigEndian.PutUint32(buf[8:12], obj.Size)
	binary.BigEndian.PutUint32(buf[12:], obj.CRC32)
	return buf
}

func (obj *ObjEntry) FromByte(v []byte) error {
	var buf []byte
	if len(v) >= 16 {
		buf = v[:16]
	} else {
		return ErrSize
	}
	obj.Offset = binary.BigEndian.Uint64(buf[:8])
	obj.Size = binary.BigEndian.Uint32(buf[8:12])
	obj.CRC32 = binary.BigEndian.Uint32(buf[12:])
	return nil
}

func (obj *ObjEntry) Key() []byte {
	var t = obj.RealType.String() + "-"
	s := "hash" + "/" + obj.Hash.String() + "/" + plumbing.Hash(obj.PackFile).String() + "/" + t[:4]
	if obj.OHeader.Type.IsDelta() {
		s = s + "/" + obj.OHeader.Reference.String()
	}
	return []byte(s)
}

func (obj *ObjEntry) SetKey(v []byte) error {
	s := string(v)
	ss := strings.Split(s, "/")
	if obj.OHeader == nil {
		obj.OHeader = &packfile.ObjectHeader{}
	}
	switch ss[3] {
	case "comm":
		obj.RealType = plumbing.CommitObject
	case "tree":
		obj.RealType = plumbing.TreeObject
	case "tag-":
		obj.RealType = plumbing.TagObject
	case "blob":
		obj.RealType = plumbing.BlobObject
	case "ofs-":
		obj.RealType = plumbing.OFSDeltaObject
	case "ref-":
		obj.RealType = plumbing.REFDeltaObject
	default:
		return ErrKey
	}
	h, err := hex.DecodeString(ss[1])
	if err != nil {
		return err
	}
	copy(obj.Hash[:], h)
	h, err = hex.DecodeString(ss[2])
	if err != nil {
		return err
	}
	copy(obj.PackFile[:], h)
	if len(ss) == 5 {
		h, err = hex.DecodeString(ss[4])
		if err != nil {
			return err
		}
		obj.OHeader.Type = plumbing.REFDeltaObject
		copy(obj.OHeader.Reference[:], h)
	} else {
		obj.OHeader.Type = obj.RealType
	}
	return nil
}

type Entries []*ObjEntry

func (es Entries) Len() int {
	return len(es)
}

func (es Entries) Less(i, j int) bool {
	return es[i].Size < es[j].Size
}

func (es Entries) Swap(i, j int) {
	es[i], es[j] = es[j], es[i]
}
