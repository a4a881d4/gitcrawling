package objext

import (
	"bufio"
	"bytes"
	"io"
	"sync"

	"gopkg.in/src-d/go-git.v4/plumbing"
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
}

func DecodeCommit(o EncodedObject) (c *object.Commit, err error) {
	c = &object.Commit{}
	if o.Type() != plumbing.CommitObject {
		return ErrUnsupportedObject
	}

	c.Hash = o.Hash()

	reader, err := o.Reader()
	if err != nil {
		return err
	}

	r := bufPool.Get().(*bufio.Reader)
	defer bufPool.Put(r)
	r.Reset(reader)

	var message bool
	var pgpsig bool
	for {
		line, err := r.ReadBytes('\n')
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
			case headerpgp:
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
	t   plumbing.ObjectType
}

func (o *BytesObject) Type() {
	return o.t
}

func (o *BytesObject) Reader() (io.Reader, error) {
	return bytes.NewBuffer(o.buf), nil
}

func NewBytesObject(b []byte, t plumbing.ObjectType) *BytesObject {
	return &BytesObject{
		buf: b,
		t:   t,
	}
}
