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
	var s = obj.OHeader.Type.String() + "-"
	if obj.OHeader.Type.IsDelta() {
		s = s[:4] + obj.Hash.String() + "/" + obj.OHeader.Reference.String() + "/" + plumbing.Hash(obj.PackFile).String()
	} else {
		s = s[:4] + obj.Hash.String() + "/" + plumbing.Hash(obj.PackFile).String()
	}
	return []byte(s)
}

func (obj *ObjEntry) SetKey(v []byte) error {
	s := string(v)
	ss := strings.Split(s, "/")
	if obj.OHeader == nil {
		obj.OHeader = &packfile.ObjectHeader{}
	}
	switch ss[0] {
	case "comm":
		obj.OHeader.Type = plumbing.CommitObject
	case "tree":
		obj.OHeader.Type = plumbing.TreeObject
	case "tag-":
		obj.OHeader.Type = plumbing.TagObject
	case "blob":
		obj.OHeader.Type = plumbing.BlobObject
	case "ofs-":
		obj.OHeader.Type = plumbing.OFSDeltaObject
	case "ref-":
		obj.OHeader.Type = plumbing.REFDeltaObject
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
	if obj.OHeader.Type.IsDelta() {
		copy(obj.OHeader.Reference[:], h)
		h, err = hex.DecodeString(ss[3])
		if err != nil {
			return err
		}
	}
	copy(obj.PackFile[:], h)
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
