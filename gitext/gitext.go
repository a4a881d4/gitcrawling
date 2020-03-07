package gitext
import (
	"bufio"
	"bytes"
	"io"
	"os"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/format/objfile"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"

	"gopkg.in/src-d/go-git.v4/storage/filesystem/dotgit"
	"gopkg.in/src-d/go-git.v4/storage/memory"
	"gopkg.in/src-d/go-git.v4/utils/ioutil"
	// "gopkg.in/src-d/go-billy.v4"
	// "gopkg.in/src-d/go-billy.v4/osfs"
)

// type GlobeStorage struct {
// 	filesystem.Storage
// 	Local *filesystem.Storage
// }

// func NewGlobeStorage(lfs,gfs billy.Filesystem) *GlobeStorage {
// 	local := filesystem.NewStorage(lfs,cache.NewObjectLRUDefault())
// 	G     := filesystem.NewStorage(gfs,cache.NewObjectLRUDefault())
	
// 	// G.ReferenceStorage = local.ReferenceStorage
// 	// G.IndexStorage     = local.IndexStorage
// 	G.ShallowStorage   = local.ShallowStorage
// 	G.ConfigStorage	   = local.ConfigStorage
// 	G.ModuleStorage    = local.ModuleStorage
// 	return &GlobeStorage{*G,local}
// }

// func (s *GlobeStorage) Init() error {
// 	return s.Local.Init()
// }

func PlainClone(url string) (*git.Repository, error) {
	o := &git.CloneOptions{
		URL:               url,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Progress: os.Stdout,
	}

	storage := memory.NewStorage()
	return git.Clone(storage,nil,o)
}

func PlainCloneFS(url,path string) (*git.Repository, error) {
	o := &git.CloneOptions{
		URL:               url,
		Depth: 1,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Progress: os.Stdout,
	}
	
	token := os.Getenv("GITHUB_AUTH_TOKEN")
	if token != "" {
		auth  := http.TokenAuth{Token:token}
		o.Auth = &auth
	}
	return git.PlainClone(path,true,o)
}
func SetEncodedObject(dir *dotgit.DotGit,o plumbing.EncodedObject) (h plumbing.Hash, err error) {
	if o.Type() == plumbing.OFSDeltaObject || o.Type() == plumbing.REFDeltaObject {
		return plumbing.ZeroHash, plumbing.ErrInvalidType
	}

	ow, err := dir.NewObject()
	if err != nil {
		return plumbing.ZeroHash, err
	}

	defer ioutil.CheckClose(ow, &err)

	or, err := o.Reader()
	if err != nil {
		return plumbing.ZeroHash, err
	}

	defer ioutil.CheckClose(or, &err)

	if err = ow.WriteHeader(o.Type(), o.Size()); err != nil {
		return plumbing.ZeroHash, err
	}

	if _, err = io.Copy(ow, or); err != nil {
		return plumbing.ZeroHash, err
	}

	return o.Hash(), err
}

func EncodedObject(o plumbing.EncodedObject) (h plumbing.Hash, blob []byte, err error) {
	if o.Type() == plumbing.OFSDeltaObject || o.Type() == plumbing.REFDeltaObject {
		return plumbing.ZeroHash,[]byte{},plumbing.ErrInvalidType
	}

	b  := bytes.NewBuffer(make([]byte, 0))
	bw := bufio.NewWriter(b)
	ow := objfile.NewWriter(bw)

	defer ioutil.CheckClose(ow, &err)

	or, err := o.Reader()
	if err != nil {
		return plumbing.ZeroHash,[]byte{}, err
	}

	defer ioutil.CheckClose(or, &err)

	if err = ow.WriteHeader(o.Type(), o.Size()); err != nil {
		return plumbing.ZeroHash,[]byte{}, err
	}

	if _, err = io.Copy(ow, or); err != nil {
		return plumbing.ZeroHash,[]byte{}, err
	}

	return o.Hash(),b.Bytes(),err
}