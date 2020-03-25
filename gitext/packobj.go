package gitext

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"

	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/format/idxfile"
	"gopkg.in/src-d/go-git.v4/utils/binary"
)

type ObjEntry struct {
	Offset, Size uint64
}

type Classifer interface {
	Hit(plumbing.Hash) bool
	FileNamePrefix() string
}
type ByteClassify byte

func (b ByteClassify) Hit(h plumbing.Hash) bool {
	return h[0] == byte(b)
}

func (b ByteClassify) FileNamePrefix() string {
	return fmt.Sprintf("pack-%02x-", b)
}

const (
	isO64Mask = uint64(1) << 31
	noMapping = -1
)

type SplitIdx []Classifer

func (s SplitIdx) FileNamePrefix(k int) (string, error) {
	if k >= len(s) {
		return "", fmt.Errorf("Out of order")
	}
	return s[k].FileNamePrefix(), nil
}

func (s SplitIdx) Number(hash plumbing.Hash) int {
	for k, v := range s {
		if v.Hit(hash) {
			return k
		}
	}
	return -1
}

func DefaultByteSplit() SplitIdx {
	r := make([]Classifer, 256)
	for k, _ := range r {
		r[k] = ByteClassify(k)
	}
	return r
}

func (s SplitIdx) GetOffset(idxf string) (objs [][]ObjEntry, err error) {
	objs = make([][]ObjEntry, len(s))

	var idx *idxfile.MemoryIndex
	idx, err = NewIdx(idxf)
	if err != nil {
		return
	}
	var pfSize uint64
	stat, err := os.Stat(strings.Replace(idxf, ".idx", ".pack", -1))
	if err != nil {
		return
	}
	pfSize = uint64(stat.Size())

	iter, err := idx.EntriesByOffset()
	if err != nil {
		return
	}
	var last, e *idxfile.Entry
	last = nil
	var Add = func(size uint64) {
		if last == nil {
			return
		}

		ma := s.Number(last.Hash)
		if ma == -1 {
			return
		}

		objs[ma] = append(objs[ma], ObjEntry{last.Offset, size})
	}

	for {
		e, err = iter.Next()
		if err == io.EOF {
			Add(pfSize - 20 - last.Offset)
			err = nil
			break
		}
		if err != nil {
			return
		}
		if last != nil {
			Add(e.Offset - last.Offset)
		}
		last = e
	}

	return
}

type PackFile struct {
	f      *os.File
	w      io.Writer
	hasher plumbing.Hasher
}

func NewPack(fn string) (*PackFile, error) {
	w, err := os.Create(fn)
	if err != nil {
		return nil, err
	}
	h := plumbing.Hasher{
		Hash: sha1.New(),
	}
	mw := io.MultiWriter(w, h)
	return &PackFile{
		f:      w,
		w:      mw,
		hasher: h,
	}, nil
}

var (
	signature  = []byte{'P', 'A', 'C', 'K'}
	ErrOutSize = errors.New("Out of Size")
)

const (
	// VersionSupported is the packfile version supported by this package
	VersionSupported uint32 = 2
)

func (pf *PackFile) Head(numEntries int) error {
	return binary.Write(
		pf.w,
		signature,
		int32(VersionSupported),
		int32(numEntries),
	)
}

func (pf *PackFile) Footer() (plumbing.Hash, error) {
	h := pf.hasher.Sum()
	return h, binary.Write(pf.w, h)
}

func (pf *PackFile) Do(r io.Reader, size int64) (int64, error) {
	return io.CopyN(pf.w, r, size)
}

func (pf *PackFile) Close() error {
	return pf.f.Close()
}

type MergeFile struct {
	dir   string
	w     *PackFile
	files map[string]uint32
	fsize uint64
}

func (m *MergeFile) AddFile(fn string, limit uint64) error {
	stat, err := os.Stat(fn)
	if err != nil {
		return err
	}
	m.fsize += uint64(stat.Size())
	f, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer f.Close()

	_, _ = binary.ReadUint32(f)
	_, _ = binary.ReadUint32(f)
	num, err := binary.ReadUint32(f)
	if err != nil {
		return err
	}
	fmt.Println("Add file", fn, num, m.fsize)

	m.files[fn] = num
	if m.fsize > limit {
		return ErrOutSize
	} else {
		return nil
	}
}
func (m *MergeFile) Flush() error {
	type withError func() error

	var step = []withError{m.Head, m.Do, m.Footer}

	for _, v := range step {
		if err := v(); err != nil {
			return err
		}
	}
	return nil
}
func (m *MergeFile) Head() error {
	var num int = 0
	for _, v := range m.files {
		num += int(v)
	}
	return m.w.Head(num)
}

func (m *MergeFile) Do() error {
	for k, _ := range m.files {
		stat, err := os.Stat(k)
		if err != nil {
			return err
		}
		size := stat.Size() - 12 - 20
		r, err := os.Open(k)
		if err != nil {
			return err
		}
		_, err = r.Seek(12, 0)
		if err != nil {
			return err
		}
		_, err = m.w.Do(r, size)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *MergeFile) Footer() error {
	hash, err := m.w.Footer()
	if err != nil {
		return err
	}
	err = m.Close()
	if err != nil {
		return err
	}
	newfn := path.Join(m.dir, "pack-"+hash.String()+".pack")
	err = os.Rename(path.Join(m.dir, "tmp-pack"), newfn)
	if err != nil {
		return err
	}
	cmd := exec.Command("git", "index-pack", newfn)
	err = cmd.Run()
	if err != nil {
		return err
	}
	fmt.Println("Flush", hash.String())
	return err
}
func (m *MergeFile) Close() error {
	return m.w.Close()
}

func NewMergeFile(dir string) (*MergeFile, error) {
	pf, err := NewPack(path.Join(dir, "tmp-pack"))
	if err != nil {
		return nil, err
	}
	f := make(map[string]uint32)
	return &MergeFile{
		dir:   dir,
		w:     pf,
		files: f,
	}, nil
}
