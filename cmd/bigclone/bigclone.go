package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/a4a881d4/gitcrawling/gitext"
)

var (
	argReposDir = flag.String("r", ".", "The dir story Repos")
	argMissDir = flag.String("m", ".", "The miss file dir")
	argThread   = flag.Int("t", 0, "Multi thread clone")
)

var (
	githubServer = []string{"github.com","github.com.cnpmjs.org"}
)
var (
	token chan int
	done  int
	all   int
	wg    sync.WaitGroup
)

func main() {
	flag.Parse()

	token = make(chan int, *argThread)

	var doSome = func(names []string, numT int) {
		defer func() {
			<-token
			fmt.Println("Done")
			wg.Done()
		}()

		for num, name := range names {
			all++
			url, path, err := GetUrlPath(name)
			if err != nil {
				fmt.Println(err)
				fmt.Printf("%06d %s bad\n", all,name)
				continue
			}
			_, err = os.Stat(path)
			if err == nil {
				fmt.Printf("%06d %s exist\n", all,name)
				continue
			}
			done++
			fmt.Printf("%5d ", numT)
			fmt.Println("Begin to Clone", url, num, done, all, time.Now())
			_, err = gitext.PlainCloneFS(url, path)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Printf("%5d ", numT)
			fmt.Println("End ", url, num, done, all, time.Now())
		}
	}
	batchDo(doSome)
	fmt.Println("Wait Clone finish")
	wg.Wait()
}

func batchDo(putSome func([]string, int)) {
	missfile := *argMissDir + "/miss"
	_, err := os.Stat(missfile)
	if err != nil {
		buildMiss(missfile)
	} else {
		updateMiss(missfile)
	}
	buf, err := ioutil.ReadFile(missfile)
	names := strings.Split(string(buf), "\n")

	var batch []string
	var numT = 0
	for _, name := range names {
		batch = append(batch, name)
		if len(batch) > 32 {
			go putSome(batch, numT)
			numT++
			wg.Add(1)
			token <- 1
			batch = []string{}
		}
	}

	if len(batch) > 0 {
		go putSome(batch, numT)
		wg.Add(1)
		token <- 1
	}
}

func GetUrlPath(name string) (url, path string, err error) {
	repo := strings.Split(name, "/")
	if len(repo) != 2 {
		url, path = "", ""
		err = fmt.Errorf("error name: %s", name)
		return
	}
	owner, project := repo[0], repo[1]

	url = fmt.Sprintf("http://%s/%s/%s", githubServer[rand.Intn(2)], owner, project)

	var bowner string
	if len(owner) > 2 {
		bowner = owner[:2] + "/" + owner
	} else {
		bowner = owner + "/" + owner
	}
	path = fmt.Sprintf("%s/repos/%s/%s", *argReposDir, bowner, project)
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
		_, path, err := GetUrlPath(name)
		if err != nil {
			continue
		}
		_, err = os.Stat(path)
		if err != nil {
			fmt.Fprintln(w, name)
		}
	}
}
