package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/a4a881d4/gitcrawling/badgerdb"
	"github.com/a4a881d4/gitcrawling/gitext"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

var (
	argReposDir = flag.String("r", ".", "The dir story Repos")
	argDir      = flag.String("t", "../.gitdb", "The dir story Trees")
	argDump     = flag.Bool("d", false, "dump Trees")
	argMode     = flag.String("m", "raw", "raw or flat")
	repoNum     = 0
	missNum     = 0
)

func main() {
	flag.Parse()
	switch *argMode {
	case "raw":
		Raw()
	case "db":
		ReadDB(flag.Arg(0))
	default:
		Flat()
	}

}
func ReadDB(hash string) {
	h, err := hex.DecodeString(hash)
	if err != nil {
		fmt.Println(err)
		return
	}
	var H plumbing.Hash
	copy(H[:], h)
	tdb, err := badgerdb.NewDB(*argDir + "/trees")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tdb.Close()
	m, err := gitext.Tree(H, tdb)
	if err != nil {
		fmt.Println(err)
		return
	}
	for k, v := range m {
		fmt.Println(gitext.TreeEntry2String(&object.TreeEntry{
			Hash: k,
			Name: v.Name,
			Mode: v.Mode,
		}))
	}
}

func Raw() {
	var err error

	tdb, err := badgerdb.NewDB(*argDir + "/trees")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tdb.Close()
	if *argDump {
		tdb.RawTrees(func(k, v []byte) error {
			fmt.Println(hex.EncodeToString(k[1:]))
			entries, err := gitext.DumpTree(v)
			if err != nil {
				fmt.Println(err)
			}
			for _, e := range entries {
				fmt.Println(gitext.TreeEntry2String(e))
			}
			return nil
		})
		return
	}

	buf, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Println(err)
		return
	}

	var rec = make(map[string][]string)
	err = json.Unmarshal(buf, &rec)
	if err != nil {
		fmt.Println(err)
		return
	}

	var putSome = func(names []string) {
		tdb.NewSession()
		defer tdb.EndSession()
		for _, name := range names {
			repo := strings.Split(name, "/")
			if len(repo) != 2 {
				fmt.Println("error name", name)
				continue
			}
			owner, project := repo[0], repo[1]
			path := fmt.Sprintf("%s/repos/%s/%s", *argReposDir, owner, project)
			_, err := os.Stat(path)
			if err != nil {
				fmt.Println(ShowName(owner, project), "miss")
				missNum++
			} else {
				path := fmt.Sprintf("%s/repos/%s/%s", *argReposDir, owner, project)
				fmt.Println(ShowName(owner, project), "check", path)
				r, err := gitext.PlainOpen(path)
				if err != nil {
					fmt.Println(ShowName(owner, project), err)
				} else {
					err = gitext.Trees(r, tdb.PutRawTree)
					if err != nil {
						fmt.Println(err)
					}
				}
			}
			repoNum++
		}
	}
	var batch []string
	for _, v := range rec {
		for _, name := range v {
			batch = append(batch, name)
		}
		if len(batch) > 2048 {
			putSome(batch)
			batch = []string{}
		}
	}
	if len(batch) > 0 {
		putSome(batch)
	}
	fmt.Printf("%8d/%d\n", repoNum-missNum, repoNum)
}

func Flat() {
	var err error

	tdb, err := badgerdb.NewDB(*argDir + "/trees")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer tdb.Close()
	if *argDump {
		prefix := "t/"
		if flag.NArg() == 1 {
			prefix += flag.Arg(0)
		}
		tdb.ForEach([]byte(prefix), func(k, v []byte) error {
			fmt.Println(string(k[2:]))
			fmt.Println(string(v))
			return nil
		})
		return
	}

	buf, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Println(err)
		return
	}

	var rec = make(map[string][]string)
	err = json.Unmarshal(buf, &rec)
	if err != nil {
		fmt.Println(err)
		return
	}

	var putSome = func(names []string) {
		tdb.NewSession()
		defer tdb.EndSession()
		for _, name := range names {
			repo := strings.Split(name, "/")
			if len(repo) != 2 {
				fmt.Println("error name", name)
				continue
			}
			owner, project := repo[0], repo[1]
			path := fmt.Sprintf("%s/repos/%s/%s", *argReposDir, owner, project)
			_, err := os.Stat(path)
			if err != nil {
				fmt.Println(ShowName(owner, project), "miss")
				missNum++
			} else {
				fmt.Println(ShowName(owner, project), "check")
				trees, err := OpenAndTree(owner, project, *argReposDir)
				if err != nil {
					fmt.Println(ShowName(owner, project), err)
				} else {
					tdb.PutTree(owner, project, trees)
				}
			}
			repoNum++
		}
	}
	var batch []string
	for _, v := range rec {
		for _, name := range v {
			batch = append(batch, name)
		}
		if len(batch) > 2048 {
			putSome(batch)
			batch = []string{}
		}
	}
	if len(batch) > 0 {
		putSome(batch)
	}
	fmt.Printf("%8d/%d\n", repoNum-missNum, repoNum)
}

func ShowName(owner, project string) string {
	var space = "                                                                        "
	num := fmt.Sprintf("%8d ", repoNum)
	if len(owner) > 25 {
		owner = owner[:25]
	} else {
		owner = space[:25-len(owner)] + owner
	}
	if len(project) > 35 {
		project = project[:35]
	} else {
		project = project + space[:35-len(project)]
	}
	return num + owner + ":" + project
}

func OpenAndTree(owner, project, ReposDir string) (ref []string, err error) {
	path := fmt.Sprintf("%s/repos/%s/%s", ReposDir, owner, project)
	r, err := gitext.PlainOpen(path)
	if err != nil {
		return []string{}, err
	}
	return gitext.TreeFlat(r)
}

func dump(ref []gitext.Ref) {
	for k, v := range ref {
		fmt.Println(k, ":", v)
	}
}
