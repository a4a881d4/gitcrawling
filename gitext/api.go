package gitext
import (
	"fmt"
	"context"
	"os"

	"github.com/google/go-github/v29/github"
	"golang.org/x/oauth2"
)
type Client struct {
	C *github.Client
	Ctx context.Context
}

func NewGitHubClient() *Client {
	token := os.Getenv("GITHUB_AUTH_TOKEN")
	if token == "" {
		fmt.Println("Unauthorized: No token present")
	}
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	ctx := context.Background()
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	return &Client{client,ctx}
}

func(c *Client) GetRef(owner,repo string) ([]Ref,error) {

	opt := &github.ReferenceListOptions{
		ListOptions: github.ListOptions{PerPage: 10},
	}
	refs, _, err := c.C.Git.ListRefs(c.Ctx, owner, repo, opt)
	if err != nil {
		return []Ref{}, err
	}
	var r = []Ref{}
	for _,v := range refs {
		ri := Ref{*v.Ref,*v.Object.SHA}
		r = append(r,ri)
	}
	return r,nil
}
