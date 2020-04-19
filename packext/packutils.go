package packext

import (
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/a4a881d4/gitcrawling/types"
)

func DupObjectInFile(fn string) (oerr error) {
	var ow *os.File
	dir := path.Dir(fn)
	tf := path.Join(dir, "tmp-pack")
	ow, oerr = os.Create(tf)

	var mo = make(map[[32]byte]uint32)
	oerr = TravelPackFile(fn, func(offset uint32, head [32]byte, size uint32, r io.ReadSeeker) (err error) {
		var n int
		var duped bool
		_, duped = mo[head]
		if duped {
			fmt.Println("Dup @", offset)
		}
		mo[head] = offset
		if duped {
			_, err = r.Seek(int64(size), 1)
			if err != nil {
				err = fmt.Errorf("Seek %v", err)
				return
			}
		} else {
			n, err = ow.Write(head[:])
			if err != nil {
				err = fmt.Errorf("Write head %v", err)
				return
			}
			if n != 32 {
				err = fmt.Errorf("Write head %d", n)
				return
			}
			var n64 int64
			n64, err = io.CopyN(ow, r, int64(size))
			if err != nil {
				err = fmt.Errorf("Write body %v", err)
				return
			}
			if n64 != int64(size) {
				err = fmt.Errorf("Write body write:%d,size:%d", n64, size)
				return
			}
		}
		return
	})
	ow.Close()
	if oerr != nil {
		return
	}
	oerr = os.Remove(fn)
	if oerr != nil {
		return oerr
	}
	oerr = os.Rename(tf, fn)
	if oerr != nil {
		return oerr
	}
	return oerr
}

func BuildIdxFile(fn string) (oerr error) {
	var ids []types.Index
	oerr = TravelPackFile(fn, func(offset uint32, head [32]byte, size uint32, r io.ReadSeeker) (err error) {
		var kp types.KeyPart
		copy(kp[:], head[1:5])
		i := types.ToIndex(kp, offset)
		ids = append(ids, i)
		_, err = r.Seek(int64(size), 1)
		if err != nil {
			err = fmt.Errorf("Seek %v", err)
			return
		}
		return
	})
	if oerr != nil {
		return
	}
	var indexes = types.Indexes(ids)
	sort.Sort(indexes)
	var ow *os.File
	tf := strings.Replace(fn, ".pack", ".idx", -1)
	ow, oerr = os.Create(tf)
	if oerr != nil {
		return
	}
	defer ow.Close()
	oerr = indexes.ToFile(ow)
	return
}
