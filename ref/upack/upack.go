package main

import (
	"fmt"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/client"
	"gopkg.in/src-d/go-git.v4/utils/ioutil"
	"gopkg.in/src-d/go-git.v4/plumbing/protocol/packp"
	"gopkg.in/src-d/go-git.v4/plumbing/protocol/packp/capability"
	"gopkg.in/src-d/go-git.v4/plumbing/protocol/packp/sideband"	
)

var (
	argUrl = flag.string("u","github.com/a4a881d4/gitcrawling","git url")
	argPackFile = flag.string("p","../temp.pack","Pack file")
)

func main() {
	flag.Parse()
	w, err := os.Create(*argPackFile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer w.Close()
	
	if err = upload(*argUrl,w); err != nil {
		fmt.Println(err)
	}
}

func upload(url string,pack io.Writer) error {
	o := &git.CloneOptions{
		URL:               url,
		Depth:             1,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Progress: os.Stdout,
	}
	ep, err := transport.NewEndpoint(o.URL)
	if err != nil {
		return  err
	}

	c, err := client.NewClient(ep)
	if err != nil {
		return err
	}

	s,err := c.NewUploadPackSession(ep, o.Auth)
	if err != nil {
		return err
	}

	defer ioutil.CheckClose(s, &err)

	ar, err := s.AdvertisedReferences()
	if err != nil {
		return err
	}

	fmt.Println(ar)
	
	req := packp.NewUploadPackRequestFromCapabilities(ar.Capabilities)
	req.Depth = packp.DepthCommits(o.Depth)

	remoteRefs, err := ar.AllReferences()
	if err != nil {
		return err
	}

	var result []plumbing.Hash
	for _, ref := range remoteRefs {
		hash := ref.Hash()
		result = append(result, h)

	req.Wants = result
	req.Haves = []plumbing.Hash{}

	reader, err := s.UploadPack(ctx, req)
	if err != nil {
		return err
	}
	defer ioutil.CheckClose(reader, &err)

	scanner := buildSidebandIfSupported(req.Capabilities, reader, o.Progress)
	io.copy(pack,scanner)

	return err
}

func buildSidebandIfSupported(l *capability.List, reader io.Reader, p sideband.Progress) io.Reader {
	var t sideband.Type

	switch {
	case l.Supports(capability.Sideband):
		t = sideband.Sideband
	case l.Supports(capability.Sideband64k):
		t = sideband.Sideband64k
	default:
		return reader
	}

	d := sideband.NewDemuxer(t, reader)
	d.Progress = p

	return d
}
