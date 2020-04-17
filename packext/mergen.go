package packext

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
)

type MergeNoFile struct {
	split SplitIdx
	dir   string
}

func (m *MergeNoFile) AddFile(idxf string) error {

	idx, err := packext.NewIdxObj(idxf)
	if err != nil {
		return err
	}
	defer idx.Close()

	var fws []*os.File
	for k, _ := range m.split {
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
		fws = append(fws, fw)
	}
	for {
		err = idx.Next(func(o *ObjEntry, h, b []byte) error {
			hs := m.split.Number(o.Hash)
			for _, n := range hs {
				var ierr error
				_, ierr = io.CopyN(fws[n], bytes.NewBuffer(o.Bytes()), 32)
				if ierr != nil {
					fmt.Println(1, ierr)
					return fmt.Errorf("%d write prefix failure: %v", n, ierr)
				}
				_, ierr = io.CopyN(fws[n], bytes.NewReader(h), int64(len(h)))
				if ierr != nil {
					return fmt.Errorf("%d write head failure: %v", n, ierr)
				}
				_, ierr = io.CopyN(fws[n], bytes.NewReader(b), int64(len(b)))
				if ierr != nil {
					return fmt.Errorf("%d write body failure: %v", n, ierr)
				}
			}
			return nil
		})
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
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
