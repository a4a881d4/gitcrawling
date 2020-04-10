package packext

import (
	"fmt"
	"path"

	"github.com/a4a881d4/gitcrawling/types"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/format/packfile"
)

type SelectFile struct {
	dir  string
	w    *PackEncodeFile
	objs []string
	g    *ObjectGet
}

func NewSelectFile(dir string, objs []string, g *ObjectGet) (*SelectFile, error) {
	pf, err := NewPack(path.Join(dir, "tmp-pack"))
	if err != nil {
		return nil, err
	}

	return &SelectFile{
		dir:  dir,
		w:    pf,
		objs: objs,
		g:    g,
	}, nil
}

func (m *SelectFile) Head() error {
	return m.w.Head(len(m.objs))
}

func (m *SelectFile) Do() (err error) {
	var c = 0
	var l = len(m.objs)
	var progress = 0
	fmt.Printf("Progress:\033[s")
	for _, hs := range m.objs {
		hash := types.Hash(plumbing.NewHash(hs))
		var head *packfile.ObjectHeader
		head, err = m.g.HeaderByHash(hash)
		if err != nil {
			return
		}
		_, err = m.w.DoHead(head)
		if err != nil {
			return
		}
		_, err = m.w.DoBody(m.g.Reader())
		if err != nil {
			return
		}
		c++
		if c > progress*l/100 {
			progress++
			fmt.Printf("\033[u\033[K(%3d%%,%d/%d)", progress, c, l)
		}
	}
	fmt.Printf("\n")
	err = nil
	return
}

func (m *SelectFile) Close() error {
	return m.w.Close()
}
func (m *SelectFile) Dir() string {
	return m.dir
}
func (m *SelectFile) Hash() (plumbing.Hash, error) {
	return m.w.Footer()
}
