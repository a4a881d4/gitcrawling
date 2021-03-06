package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	ospath "path"
	"strings"
	"sync"
	"time"

	"github.com/a4a881d4/gitcrawling/gitext"
	"gopkg.in/src-d/go-git.v4/storage/memory"
)

var (
	argReposDir = flag.String("r", ".", "The dir story Packs")
	argMissDir  = flag.String("m", ".", "The miss file dir")
	argThread   = flag.Int("t", 1, "Multi thread clone")
)

var (
	githubServer = []string{
		"git://github.com",
		"http://github.com.cnpmjs.org",
		"http://github.com.cnpmjs.org",
	}
)
var (
	token chan int
	done  int
	all   int
	wg    sync.WaitGroup
)

func getPack(url, tempf string, numT int) (refs string, err error) {
	done++
	startTime := time.Now()
	fmt.Printf("%5d ", numT)
	fmt.Println("Begin to Clone", url, done, all,
		startTime.Format("2006-01-02 15:04:05"))
	refs = ""
	w, err := os.Create(tempf)
	if err != nil {
		return
	}
	defer w.Close()

	var rmap memory.ReferenceStorage
	if rmap, err = gitext.Upload(url, w); err != nil {
		return
	}

	for k, v := range rmap {
		refs += fmt.Sprintf("%s: %s\n", k.String(), v.Hash().String())
	}

	fmt.Printf("%5d ", numT)
	endTime := time.Now()
	Duration := endTime.Sub(startTime)
	fmt.Println("End ", url, done, all,
		endTime.Format("2006-01-02 15:04:05"), Duration.Seconds())
	return
}

func clone(numT int, task chan string) {
	for {
		name := <-task
		if name == "END" {
			break
		}
		all++
		url, path, filename, err := GetUrlPath(name)
		if err != nil {
			fmt.Printf("%06d %s bad\n", all, name)
			continue
		}
		_, err = os.Stat(path)
		if err == nil {
			fmt.Printf("%06d %s exist\n", all, name)
			continue
		}
		os.MkdirAll(path, os.ModePerm)
		pf := ospath.Join(path, filename)
		tempf := ospath.Join(path, "tmp-pack")
		if refs, err := getPack(url, tempf, numT); err != nil {
			fmt.Printf("%6d ", numT)
			fmt.Println(err, path, "will be removed")
			os.RemoveAll(path)
		} else {
			ioutil.WriteFile(ospath.Join(path, "refs"), []byte(refs), 0755)
			os.Rename(tempf, pf)
		}
	}
	fmt.Println("Worker", numT, "Done")
	wg.Done()
}

func main() {
	flag.Parse()

	var task = make(chan string, *argThread)

	for i := 0; i < *argThread; i++ {
		go clone(i, task)
	}

	wg.Add(*argThread)

	batchDo(task)

	for i := 0; i < *argThread*2; i++ {
		task <- "END"
	}

	fmt.Println("Wait Clone finish")
	wg.Wait()
}

func batchDo(task chan string) {
	missfile := *argMissDir + "/miss"
	_, err := os.Stat(missfile)
	if err != nil {
		buildMiss(missfile)
	} else {
		updateMiss(missfile)
	}
	buf, err := ioutil.ReadFile(missfile)
	names := strings.Split(string(buf), "\n")

	for _, name := range names {
		name = strings.Replace(name, "\r", "", -1)
		task <- name
	}
}

func GetUrlPath(name string) (url, path, filename string, err error) {
	name = strings.Replace(name, "\r", "", -1)
	repo := strings.Split(name, "/")
	if len(repo) != 2 {
		url, path = "", ""
		err = fmt.Errorf("error name: %s", name)
		return
	}
	owner, project := repo[0], repo[1]

	url = fmt.Sprintf("%s/%s/%s", githubServer[rand.Intn(3)], owner, project)

	var bowner string
	if len(owner) > 2 {
		bowner = owner[:2] + "/" + owner
	} else {
		bowner = owner + "/" + owner
	}
	path = fmt.Sprintf("%s/packs/%s/%s", *argReposDir, bowner, project)
	filename = fmt.Sprintf("pack-%s.%s.pack", owner, project)
	err = nil
	return
}

func buildMiss(missfile string) {
	w, err := os.Create(missfile)
	defer w.Close()
	if err != nil {
		fmt.Println(err)
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
	for _, v := range rec {
		for _, name := range v {
			fmt.Fprintln(w, name)
		}
	}
}

func updateMiss(missfile string) {
	buf, err := ioutil.ReadFile(missfile)
	names := strings.Split(string(buf), "\n")
	err = os.Remove(missfile)
	if err != nil {
		fmt.Println(err)
		return
	}
	w, err := os.Create(missfile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer w.Close()

	for _, name := range names {
		_, path, _, err := GetUrlPath(name)
		if err != nil {
			continue
		}
		_, err = os.Stat(path)
		if err != nil {
			fmt.Fprintln(w, name)
		}
	}
}
