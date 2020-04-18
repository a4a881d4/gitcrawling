package packext

import (
	"fmt"
	"os"
	"path"

	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/utils/binary"
)

type MergeFile struct {
	dir   string
	w     *PackEncodeFile
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

func (m *MergeFile) Hash() (plumbing.Hash, error) {
	return m.w.Footer()
}

func (m *MergeFile) Close() error {
	return m.w.Close()
}
func (m *MergeFile) Dir() string {
	return m.dir
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
