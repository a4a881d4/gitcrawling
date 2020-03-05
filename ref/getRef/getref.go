package main

import (
	"fmt"

	"github.com/a4a881d4/gitcrawling/gitext"
)

func main() {
	c := gitext.NewGitHubClientWithoutToken()
	_,repos,err := c.ListAll(0)
	if err!=nil {
		fmt.Println(err)
	}
	fmt.Println(repos)
	ref,err := c.GetRef("a4a881d4","gitcrawling")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(ref)
}