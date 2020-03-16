package main

import (
	"fmt"
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
