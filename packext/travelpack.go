package packext

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

func TravelPackFile(fn string, cb func(o uint32, h [32]byte, s uint32, r io.ReadSeeker) error) (err error) {
	var fr *os.File
	fr, err = os.Open(fn)
	if err != nil {
		return
	}
	defer fr.Close()
	var head [32]byte
	var offset uint32
	for err == nil {
		var n int
		n, err = fr.Read(head[:])
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			err = fmt.Errorf("Read %v", err)
			break
		}
		if n != 32 {
			err = fmt.Errorf("Read %d", n)
			break
		}
		size := binary.BigEndian.Uint32(head[20:24])
		cb(offset, head, size, fr)
		offset += size + 32
	}
	return
}
