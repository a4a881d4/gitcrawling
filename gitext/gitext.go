package gitext
import (
	"io"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
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
	}

	storage := memory.NewStorage()
	return git.Clone(storage,nil,o)
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