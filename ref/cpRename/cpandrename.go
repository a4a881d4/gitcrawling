package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	ospath "path"
	"strings"

	"github.com/a4a881d4/gitcrawling/gitext"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

var (
	argReposDir = flag.String("r", ".", "The dir story Packs")
	argToDir    = flag.String("t", ".", "copy to")
	argMode     = flag.String("m", "gen", "copy to")
)

func getHash(fn string) (hash plumbing.Hash, err error) {
	var f *os.File
	f, err = os.Open(fn)
	if err != nil {
		return
	}
	_, err = f.Seek(-20, 2)
	if err != nil {
		return
	}
	_, err = io.ReadFull(f, hash[:])
	return
}

func gen() {
	o1, err := readProject(flag.Arg(0))
	if err != nil {
		fmt.Println(err)
		return
	}
	o2, err := readProject(flag.Arg(1))
	if err != nil {
		fmt.Println(err)
		return
	}
	for k, _ := range o2 {
		if strings.Contains(k, "/") {
			if _, ok := o1[k]; ok {
				o1[k] = true
			}
		}
	}
	err = writeProject(flag.Arg(2), o1)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func readProject(fn string) (map[string]bool, error) {
	r := make(map[string]bool)
	buf, err := ioutil.ReadFile(fn)
	if err != nil {
		return r, err
	}
	names := strings.Split(string(buf), "\n")
	for _, name := range names {
		name = strings.Replace(name, "\r", "", -1)
		name = strings.TrimSpace(name)
		r[name] = false
	}
	return r, nil
}

func writeProject(fn string, m map[string]bool) error {
	f, err := os.Create(fn)
	if err != nil {
		return err
	}
	defer f.Close()
	for k, v := range m {
		if v {
			fmt.Fprintln(f, k)
		}
	}
	return nil
}

func dumpHash(cp bool) {
	o, err := readProject(flag.Arg(0))
	if err != nil {
		fmt.Println(err)
		return
	}
	for k, _ := range o {
		_, path, name, err := gitext.GetUrlPath(k, *argReposDir, []string{})
		if err != nil {
			fmt.Println(err)
			continue
		}
		fn := ospath.Join(path, name)
		_, err = os.Stat(fn)
		if err != nil {
			fmt.Println(err)
			continue
		}
		hash, err := getHash(fn)
		if err != nil {
			fmt.Println(err)
			continue
		} else {
			fmt.Println(hash.String())
		}
		if cp {
			err = cpAndIdx(fn, hash)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}

func cpAndIdx(fn string, hash plumbing.Hash) (err error) {
	newfn := ospath.Join(*argToDir, "pack-"+hash.String()+".pack")
	var old, new *os.File
	old, err = os.Open(fn)
	if err != nil {
		return
	}
	defer old.Close()
	new, err = os.Create(newfn)
	if err != nil {
		return
	}
	_, err = io.Copy(new, old)
	if err != nil {
		return
	}
	err = new.Close()
	if err != nil {
		return
	}
	cmd := exec.Command("git", "index-pack", newfn)
	err = cmd.Run()
	if err != nil {
		fmt.Println("need remove", newfn)
		os.Remove(newfn)
	}
	return
}

func main() {
	flag.Parse()
	switch *argMode {
	case "gen":
		gen()
	case "cpRename":
		dumpHash(true)
	case "dumpHash":
		dumpHash(false)
	default:
		return
	}
}
