package main

import (
	"fmt"
	"os"
	"flag"
	"github.com/a4a881d4/gitcrawling/gitext"
)

var (
	argUrl = flag.String("u","http://github.com/a4a881d4/gitcrawling.git","git url")
	argPackFile = flag.String("p","../temp/temp.pack","Pack file")
)

func main() {
	flag.Parse()
	w, err := os.Create(*argPackFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer w.Close()
	
	if err = gitext.Upload(*argUrl,w); err != nil {
		fmt.Println(err)
	}
}

// func upload(url string,pack io.Writer) error {
// 	o := &git.CloneOptions{
// 		URL:               url,
// 		Depth:             1,
// 		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
// 		Progress: os.Stdout,
// 	}
// 	ep, err := transport.NewEndpoint(o.URL)
// 	if err != nil {
// 		return  err
// 	}

// 	c, err := client.NewClient(ep)
// 	if err != nil {
// 		return err
// 	}

// 	s,err := c.NewUploadPackSession(ep, o.Auth)
// 	if err != nil {
// 		return err
// 	}

// 	defer ioutil.CheckClose(s, &err)

// 	ar, err := s.AdvertisedReferences()
// 	if err != nil {
// 		return err
// 	}

// 	fmt.Println(ar)
	
// 	req := packp.NewUploadPackRequestFromCapabilities(ar.Capabilities)
// 	req.Depth = packp.DepthCommits(o.Depth)
// 	req.Shallows = make([]plumbing.Hash, 0)
// 	if err := req.Capabilities.Set(capability.Shallow); err != nil {
// 		return err
// 	}
// 	remoteRefs, err := ar.AllReferences()
// 	if err != nil {
// 		fmt.Println(1,err)
		
// 		return err
// 	}
// 	RefSpecs := cloneRefSpec(o)
// 	refs, err := calculateRefs(RefSpecs, remoteRefs, o.Tags)
// 	if err != nil {
// 		fmt.Println(2,err)
		
// 		return err
// 	}

// 	var result []plumbing.Hash
// 	for _, ref := range refs {
// 		result = append(result, ref.Hash())
// 	}
// 	req.Wants = result
// 	req.Haves = []plumbing.Hash{}
// 	fmt.Println(req)
// 	reader, err := s.UploadPack(context.Background(), req)
// 	if err != nil {
// 		fmt.Println(3,err)
// 		return err
// 	}
// 	defer ioutil.CheckClose(reader, &err)

// 	scanner := buildSidebandIfSupported(req.Capabilities, reader, o.Progress)
// 	io.Copy(pack,scanner)

// 	return err
// }

// func buildSidebandIfSupported(l *capability.List, 
// 	reader io.Reader, p sideband.Progress) io.Reader {
// 	var t sideband.Type

// 	switch {
// 	case l.Supports(capability.Sideband):
// 		t = sideband.Sideband
// 	case l.Supports(capability.Sideband64k):
// 		t = sideband.Sideband64k
// 	default:
// 		return reader
// 	}

// 	d := sideband.NewDemuxer(t, reader)
// 	d.Progress = p

// 	return d
// }

// const refspecAllTags = "+refs/tags/*:refs/tags/*"

// func calculateRefs(
// 	spec []config.RefSpec,
// 	remoteRefs storer.ReferenceStorer,
// 	tagMode git.TagMode,
// ) (memory.ReferenceStorage, error) {
// 	if tagMode == git.AllTags {
// 		spec = append(spec, refspecAllTags)
// 	}

// 	refs := make(memory.ReferenceStorage)
// 	for _, s := range spec {
// 		if err := doCalculateRefs(s, remoteRefs, refs); err != nil {
// 			return nil, err
// 		}
// 	}

// 	return refs, nil
// }

// func doCalculateRefs(
// 	s config.RefSpec,
// 	remoteRefs storer.ReferenceStorer,
// 	refs memory.ReferenceStorage,
// ) error {
// 	iter, err := remoteRefs.IterReferences()
// 	if err != nil {
// 		return err
// 	}

// 	var matched bool
// 	err = iter.ForEach(func(ref *plumbing.Reference) error {
// 		if !s.Match(ref.Name()) {
// 			return nil
// 		}

// 		if ref.Type() == plumbing.SymbolicReference {
// 			target, err := storer.ResolveReference(remoteRefs, ref.Name())
// 			if err != nil {
// 				return err
// 			}

// 			ref = plumbing.NewHashReference(ref.Name(), target.Hash())
// 		}

// 		if ref.Type() != plumbing.HashReference {
// 			return nil
// 		}

// 		matched = true
// 		if err := refs.SetReference(ref); err != nil {
// 			return err
// 		}

// 		if !s.IsWildcard() {
// 			return storer.ErrStop
// 		}

// 		return nil
// 	})

// 	if !matched && !s.IsWildcard() {
// 		return fmt.Errorf("couldn't find remote ref %q", s.Src())
// 	}

// 	return err
// }

// func cloneRefSpec(o *git.CloneOptions) []config.RefSpec {
// 	return []config.RefSpec{
// 		config.RefSpec(fmt.Sprintf(config.DefaultFetchRefSpec, o.RemoteName)),
// 	}
// }