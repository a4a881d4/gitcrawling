package packext

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
	"sync"

	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/format/packfile"
	"gopkg.in/src-d/go-git.v4/utils/ioutil"
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

var (
	zlibInitBytes = []byte{0x78, 0x9c, 0x01, 0x00, 0x00, 0xff, 0xff, 0x00, 0x00, 0x00, 0x01}
)

type Scanner struct {
	RawObj []byte
	pos    int
}

func NewScanner(o []byte) *Scanner {
	return &Scanner{
		RawObj: o,
	}
}
func (s *Scanner) ReadByte() (byte, error) {
	if s.pos >= len(s.RawObj) {
		return byte(0), io.EOF
	}
	c := s.RawObj[s.pos]
	s.pos++
	return c, nil
}
func (s *Scanner) Reader() io.Reader {
	return bytes.NewReader(s.RawObj[s.pos:])
}
func (s *Scanner) readHash() (plumbing.Hash, error) {
	var h plumbing.Hash
	if s.pos+20 > len(s.RawObj) {
		return plumbing.ZeroHash, io.EOF
	}
	copy(h[:], s.RawObj[s.pos:s.pos+20])
	s.pos += 20
	return h, nil
}
func (s *Scanner) readVariableWidthInt() (int64, error) {
	var c byte
	var err error
	if c, err = s.ReadByte(); err != nil {
		return 0, err
	}

	var v = int64(c & maskLength)
	for c&maskContinue > 0 {
		v++
		if c, err = s.ReadByte(); err != nil {
			return 0, err
		}

		v = (v << lengthBits) + int64(c&maskLength)
	}

	return v, nil
}
func (s *Scanner) readObjectTypeAndLength() (plumbing.ObjectType, int64, error) {
	t, c, err := s.readType()
	if err != nil {
		return t, 0, err
	}

	l, err := s.readLength(c)

	return t, l, err
}
func (s *Scanner) Header() (*packfile.ObjectHeader, error) {
	var err error
	h := &packfile.ObjectHeader{}

	h.Type, h.Length, err = s.readObjectTypeAndLength()
	if err != nil {
		return nil, err
	}

	switch h.Type {
	case plumbing.OFSDeltaObject:
		no, err := s.readVariableWidthInt()
		if err != nil {
			return nil, err
		}

		h.OffsetReference = h.Offset - no
	case plumbing.REFDeltaObject:
		var err error
		h.Reference, err = s.readHash()
		if err != nil {
			return nil, err
		}
	}

	return h, nil
}

func (s *Scanner) readType() (plumbing.ObjectType, byte, error) {
	var c byte
	var err error
	if c, err = s.ReadByte(); err != nil {
		return plumbing.ObjectType(0), 0, err
	}

	typ := parseType(c)

	return typ, c, nil
}

func parseType(b byte) plumbing.ObjectType {
	return plumbing.ObjectType((b & maskType) >> firstLengthBits)
}

// the length is codified in the last 4 bits of the first byte and in
// the last 7 bits of subsequent bytes.  Last byte has a 0 MSB.
func (s *Scanner) readLength(first byte) (int64, error) {
	length := int64(first & maskFirstLength)

	c := first
	shift := firstLengthBits
	var err error
	for c&maskContinue > 0 {
		if c, err = s.ReadByte(); err != nil {
			return 0, err
		}

		length += int64(c&maskLength) << shift
		shift += lengthBits
	}

	return length, nil
}

var zlibReaderPool = sync.Pool{
	New: func() interface{} {
		r, _ := zlib.NewReader(bytes.NewReader(zlibInitBytes))
		return r
	},
}
var byteSlicePool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(nil)
	},
}

func (s *Scanner) copyObject(w io.Writer) (n int64, err error) {
	zr := zlibReaderPool.Get().(io.ReadCloser)
	defer zlibReaderPool.Put(zr)
	sr := s.Reader()
	if err = zr.(zlib.Resetter).Reset(sr, nil); err != nil {
		return 0, fmt.Errorf("zlib reset error: %s", err)
	}

	defer ioutil.CheckClose(zr, &err)
	buf := byteSlicePool.Get().([]byte)
	n, err = io.CopyBuffer(w, zr, buf)
	byteSlicePool.Put(buf)
	return
}
