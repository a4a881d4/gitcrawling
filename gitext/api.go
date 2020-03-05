package gitext
import (
	"context"
	"fmt"
	"os"
	"time"

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
func NewGitHubClientWithoutToken() *Client {
	
	ctx := context.Background()
	client := github.NewClient(nil)
	return &Client{client,ctx}
}
func(c *Client) GetRef(owner,repo string) ([]Ref,error) {

	opt := &github.ReferenceListOptions{
		ListOptions: github.ListOptions{PerPage: 10},
	}
	refs, resp, err := c.C.Git.ListRefs(c.Ctx, owner, repo, opt)
	if err != nil {
		return []Ref{}, err
	}
	if resp.Rate.Remaining<=1 {
		<- time.After(time.Until(resp.Rate.Reset))
	}
	var r = []Ref{}
	for _,v := range refs {
		ri := Ref{*v.Ref,*v.Object.SHA}
		r = append(r,ri)
	}
	return r,nil
}

type Repository struct {
	ID int64
	FullName string
}

func(c *Client) ListAll(start int64) (int64,[]Repository, error) {
	opts := &github.RepositoryListAllOptions{start}
	repos,_,err := c.C.Repositories.ListAll(c.Ctx, opts)
	if err != nil {
		return start+1,[]Repository{},err
	}
	var r = []Repository{}
	for _,v := range repos {
		ri := Repository{*v.ID,*v.FullName}
		r = append(r,ri)
	}
	return r[len(r)-1].ID,r,nil
}