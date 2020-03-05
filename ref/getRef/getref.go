package main

import (
	"fmt"

	"github.com/a4a881d4/gitcrawling/gitext"
)

func main() {
	c := gitext.NewGitHubClient()
	ref,err := c.GetRef("a4a881d4","gitcrawling")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(ref)
}