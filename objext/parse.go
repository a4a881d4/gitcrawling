package objext

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/a4a881d4/gitcrawling/packext"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/filemode"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

var bufPool = sync.Pool{
	New: func() interface{} {
		return bufio.NewReader(nil)
	},
}

type EncodedObject interface {
	Reader() (io.Reader, error)
	Type() plumbing.ObjectType
	Hash() plumbing.Hash
	Size() int64
}

func Tree2String(t *object.Tree) string {
	r := ""
	for _, v := range t.Entries {
		r += fmt.Sprintln(v.Name, v.Hash.String(), v.Mode.String())
	}
	return r
}
func DecodeTree(o EncodedObject) (t *object.Tree, err error) {
	t = &object.Tree{}
	if o.Type() != plumbing.TreeObject {
		err = object.ErrUnsupportedObject
		return
	}

	t.Hash = o.Hash()
	if o.Size() == 0 {
		return
	}

	t.Entries = nil
	var reader io.Reader
	reader, err = o.Reader()
	if err != nil {
		return
	}

	r := bufPool.Get().(*bufio.Reader)
	defer bufPool.Put(r)
	r.Reset(reader)
	for {
		var str string
		str, err = r.ReadString(' ')
		if err != nil {
			if err == io.EOF {
				break
			}

			return
		}
		str = str[:len(str)-1] // strip last byte (' ')
		var mode filemode.FileMode
		mode, err = filemode.New(str)
		if err != nil {
			return
		}
		var name string
		name, err = r.ReadString(0)
		if err != nil && err != io.EOF {
			return
		}

		var hash plumbing.Hash
		if _, err = io.ReadFull(r, hash[:]); err != nil {
			return
		}

		baseName := name[:len(name)-1]
		t.Entries = append(t.Entries, object.TreeEntry{
			Hash: hash,
			Mode: mode,
			Name: baseName,
		})
	}

	return
}

func DecodeCommit(o EncodedObject) (c *object.Commit, err error) {
	c = &object.Commit{}
	if o.Type() != plumbing.CommitObject {
		err = object.ErrUnsupportedObject
		return
	}

	c.Hash = o.Hash()
	var reader io.Reader
	reader, err = o.Reader()
	if err != nil {
		return
	}

	r := bufPool.Get().(*bufio.Reader)
	defer bufPool.Put(r)
	r.Reset(reader)

	var message bool
	var pgpsig bool
	for {
		var line []byte
		line, err = r.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return
		}

		if pgpsig {
			if len(line) > 0 && line[0] == ' ' {
				line = bytes.TrimLeft(line, " ")
				c.PGPSignature += string(line)
				continue
			} else {
				pgpsig = false
			}
		}

		if !message {
			line = bytes.TrimSpace(line)
			if len(line) == 0 {
				message = true
				continue
			}

			split := bytes.SplitN(line, []byte{' '}, 2)

			var data []byte
			if len(split) == 2 {
				data = split[1]
			}

			switch string(split[0]) {
			case "tree":
				c.TreeHash = plumbing.NewHash(string(data))
			case "parent":
				c.ParentHashes = append(c.ParentHashes, plumbing.NewHash(string(data)))
			case "author":
				c.Author.Decode(data)
			case "committer":
				c.Committer.Decode(data)
			case "gpgsig":
				c.PGPSignature += string(data) + "\n"
				pgpsig = true
			}
		} else {
			c.Message += string(line)
		}

		if err == io.EOF {
			err = nil
			return
		}
	}
	return
}

type BytesObject struct {
	buf []byte
	oe  *packext.ObjEntry
}

func (o *BytesObject) Size() int64 {
	return int64(o.oe.Size)
}
func (o *BytesObject) Type() plumbing.ObjectType {
	return o.oe.RealType
}
func (o *BytesObject) Hash() plumbing.Hash {
	return o.oe.Hash
}
func (o *BytesObject) Reader() (io.Reader, error) {
	return bytes.NewBuffer(o.buf), nil
}

func NewBytesObject(b []byte, oe *packext.ObjEntry) *BytesObject {
	return &BytesObject{
		buf: b,
		oe:  oe,
	}
}
