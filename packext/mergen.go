package packext

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

type MergeNoFile struct {
	split SplitIdx
	dir   string
}

func (m *MergeNoFile) AddFile(idxf string) error {

	objs, err := m.split.GetOffset(idxf)
	if err != nil {
		return err
	}
	pf := strings.Replace(idxf, ".idx", ".pack", -1)
	fr, err := os.Open(pf)
	if err != nil {
		return err
	}
	defer fr.Close()
	for k, v := range objs {
		if len(v) == 0 {
			continue
		}
		fnb, err := m.split.FileNamePrefix(k)
		if err != nil {
			return err
		}
		fn := path.Join(m.dir, fnb+"no.pack")
		var fw *os.File
		if _, err := os.Stat(fn); err != nil {
			fw, err = os.Create(fn)
			if err != nil {
				fmt.Println(0, err)
				return err
			}
		} else {
			fw, err = os.OpenFile(fn, os.O_APPEND|os.O_RDWR, 0644)
			if err != nil {
				fmt.Println(-1, err)
				return err
			}
		}

		defer fw.Close()
		for _, o := range v {
			_, err := io.CopyN(fw, bytes.NewBuffer(o.Bytes()), 32)
			if err != nil {
				fmt.Println(1, err)
				return err
			}
			_, err = fr.Seek(int64(o.Offset), 0)
			if err != nil {
				fmt.Println(2, err)
				return err
			}
			_, err = io.CopyN(fw, fr, int64(o.Size))
			if err != nil {
				fmt.Println(3, err)
				return err
			}
		}
	}
	return nil
}

func NewMergeNoFile(dir string) (*MergeNoFile, error) {

	return &MergeNoFile{
		dir:   dir,
		split: DefaultByteSplit(),
	}, nil
}
