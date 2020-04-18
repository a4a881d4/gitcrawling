package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
)

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

func batchDo(putSome func([]string)) {
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
}
