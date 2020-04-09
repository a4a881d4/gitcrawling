package packext

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"

	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/format/packfile"
	"gopkg.in/src-d/go-git.v4/utils/binary"
)

type PackEncodeFile struct {
	f      *os.File
	w      io.Writer
	hasher plumbing.Hasher
}

func NewPack(fn string) (*PackEncodeFile, error) {
	w, err := os.Create(fn)
	if err != nil {
		return nil, err
	}
	h := plumbing.Hasher{
		Hash: sha1.New(),
	}
	mw := io.MultiWriter(w, h)
	return &PackEncodeFile{
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

	firstLengthBits = uint8(4)   // the first byte into object header has 4 bits to store the length
	lengthBits      = uint8(7)   // subsequent bytes has 7 bits to store the length
	maskFirstLength = 15         // 0000 1111
	maskContinue    = 0x80       // 1000 0000
	maskLength      = uint8(127) // 0111 1111
	maskType        = uint8(112) // 0111 0000
)

func skipHead(buf []byte) []byte {
	for (buf[0]&maskContinue) == maskContinue && len(buf) > 1 {
		buf = buf[1:]
	}
	return buf[1:]
}
func (pf *PackEncodeFile) Head(numEntries int) error {
	return binary.Write(
		pf.w,
		signature,
		int32(VersionSupported),
		int32(numEntries),
	)
}

func (pf *PackEncodeFile) Footer() (plumbing.Hash, error) {
	h := pf.hasher.Sum()
	return h, binary.Write(pf.w, h)
}

func (pf *PackEncodeFile) Do(r io.Reader, size int64) (int64, error) {
	return io.CopyN(pf.w, r, size)
}
func (pf *PackEncodeFile) DoBody(r io.Reader) (int64, error) {
	return io.Copy(pf.w, r)
}
func (pf *PackEncodeFile) DoHead(oh *packfile.ObjectHeader) (int64, error) {
	buf := entryHead(oh)
	return io.Copy(pf.w, bytes.NewBuffer(buf))
}
func entryHead(oh *packfile.ObjectHeader) []byte {
	t := int64(oh.Type)
	size := oh.Length
	header := []byte{}
	c := (t << firstLengthBits) | (size & maskFirstLength)
	size >>= firstLengthBits
	for {
		if size == 0 {
			break
		}
		header = append(header, byte(c|maskContinue))
		c = size & int64(maskLength)
		size >>= lengthBits
	}

	header = append(header, byte(c))
	if oh.Type.IsDelta() {
		header = append(header, oh.Reference[:]...)
	}
	return header
}

func (pf *PackEncodeFile) Close() error {
	return pf.f.Close()
}

type Packer interface {
	Head() error
	Do() error
	Close() error
	Hash() (plumbing.Hash, error)
	Dir() string
}

func Flush(m Packer) error {
	type withError func() error

	var step = []withError{m.Head, m.Do, func() error { return Footer(m) }}

	for _, v := range step {
		if err := v(); err != nil {
			return err
		}
	}
	return nil
}

func Footer(m Packer) error {
	hash, err := m.Hash()
	if err != nil {
		return err
	}
	err = m.Close()
	if err != nil {
		return err
	}
	newfn := path.Join(m.Dir(), "pack-"+hash.String()+".pack")
	err = os.Rename(path.Join(m.Dir(), "tmp-pack"), newfn)
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
