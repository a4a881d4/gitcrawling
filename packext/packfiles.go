package packext

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/a4a881d4/gitcrawling/types"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

type FileFinder interface {
	Hash2FileName(types.Hash) (string, error)
}

type PackFile struct {
	H    *os.File
	Lock sync.Mutex
	Hash types.Hash
	Hit  int
}

func NewPackFileFromFN(fn string, h types.Hash) (*PackFile, error) {
	H, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	return &PackFile{
		H:    H,
		Hash: h,
	}, nil
}
func (ps *PackFiles) NewPackFile(h types.Hash) (*PackFile, error) {
	fn, err := ps.GetFileName.Hash2FileName(h)
	if err != nil {
		return nil, err
	}
	return NewPackFileFromFN(fn, h)
}

func (pf *PackFile) Close() error {
	return pf.H.Close()
}

func (pf *PackFile) Get(e *ObjEntry) ([]byte, error) {
	pf.Lock.Lock()
	defer pf.Lock.Unlock()
	_, err := pf.H.Seek(int64(e.Offset), 0)
	if err != nil {
		return []byte{}, err
	}
	r := make([]byte, e.Size)
	_, err = io.ReadFull(pf.H, r)
	return r, err
}

type PackFiles struct {
	Opened      map[types.Hash]*PackFile
	MaxOpen     int
	GetFileName FileFinder
}

func NewPackFiles(finder FileFinder) *PackFiles {
	if finder == nil {
		finder = DefaultMap
	}
	return &PackFiles{
		Opened:      make(map[types.Hash]*PackFile),
		GetFileName: finder,
	}
}

func (ps *PackFiles) Get(e *ObjEntry) ([]byte, error) {
	var h types.Hash
	h = types.Hash(e.PackFile)
	if _, ok := ps.Opened[h]; !ok {
		pf, err := ps.NewPackFile(h)
		if err != nil {
			return []byte{}, err
		}
		ps.Opened[h] = pf
	}

	if pf, ok := ps.Opened[h]; ok {
		return pf.Get(e)
	} else {
		return []byte{}, fmt.Errorf("Miss Pack File %s", plumbing.Hash(h).String())
	}
}

type FileMap map[types.Hash]string

var (
	DefaultMap FileMap = make(map[types.Hash]string)
)

func (m FileMap) Hash2FileName(h types.Hash) (string, error) {
	if fn, ok := m[h]; ok {
		return fn, nil
	} else {
		return "", fmt.Errorf("Miss Pack File %s", plumbing.Hash(h).String())
	}
}
func DefaultFromDir(dir string) error {
	stat, err := os.Stat(dir)
	if err != nil {
		return err
	}
	if stat.IsDir() {
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			fn := filepath.Base(path)
			if strings.Contains(fn, ".pack") && strings.Contains(fn, "pack-") {
				if len(fn) > 45 {
					key := fn[5:45]
					h, err := hex.DecodeString(key)
					if err == nil {
						var hash types.Hash
						copy(hash[:], h[:])
						DefaultMap[hash] = path
					}
				}
			}
			return nil
		})
	} else {
		return fmt.Errorf("%s must be dir", dir)
	}
	return nil
}
