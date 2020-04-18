package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/a4a881d4/gitcrawling/types"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()
	test, terr := hex.DecodeString(os.Args[2])
	var key types.KeyPart
	if terr == nil {
		copy(key[:], test[:])
	}

	pf, err := os.Open(strings.Replace(os.Args[1], ".idx", ".pack", -1))
	if err != nil {
		fmt.Println(err)
	}
	defer pf.Close()
	ids, err := types.IndexFromFile(f)
	if err != nil {
		fmt.Println(err)
	}
	for k, i := range ids {
		_, off := types.FromIndex(i)
		pf.Seek(int64(off), 0)
		dump(pf, k)
	}
	if terr == nil {
		p := ids.Find(key)
		if len(p) == 0 {
			fmt.Printf("no find pos: %x\n", key)
		} else {
			for _, v := range p {
				pf.Seek(int64(v), 0)
				dump(pf, 0)
			}
		}
	}
}

func dump(pf *os.File, k int) {
	var err error
	var head, base [32]byte
	var ihead [5]byte
	var spaces = "         "
	var size uint32
	pf.Read(head[:])
	pf.Read(ihead[:])
	t := plumbing.ObjectType(ihead[0] & 0xf)
	size = binary.BigEndian.Uint32(head[20:24])
	size -= 5
	ts := t.String()
	if ihead[0]&0x10 != 0 {
		pf.Read(base[:])
		size -= 32
		ts = "d " + ts
	}
	ts = spaces[:len(spaces)-len(ts)] + ts
	fmt.Printf("%10d %040x-%04x-%04x-%04x %s\n", k, head[:20], head[20:24], head[24:28], head[28:], ts)
	buf := new(bytes.Buffer)
	buf.Reset()
	var zlibInitBytes = []byte{0x78, 0x9c, 0x01, 0x00, 0x00, 0xff, 0xff, 0x00, 0x00, 0x00, 0x01}
	zr, _ := zlib.NewReader(bytes.NewReader(zlibInitBytes))
	if err = zr.(zlib.Resetter).Reset(pf, nil); err != nil {
		fmt.Printf("zlib reset error: %s", err)
	}

	bufs := make([]byte, 32*1024)
	_, err = io.CopyBuffer(os.Stdout, zr, bufs)
	os.Stdout.Sync()
	zr.Close()
	fmt.Println()

}
