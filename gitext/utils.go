package gitext

import (
	"fmt"
	"math/rand"
	"strings"
)

func GetUrlPath(name, ReposDir string, githubServer []string) (url, path, filename string, err error) {
	name = strings.Replace(name, "\r", "", -1)
	repo := strings.Split(name, "/")
	if len(repo) != 2 {
		url, path = "", ""
		err = fmt.Errorf("error name: %s", name)
		return
	}
	owner, project := repo[0], repo[1]
	if len(githubServer) != 0 {
		url = fmt.Sprintf("%s/%s/%s", githubServer[rand.Intn(len(githubServer))], owner, project)
	} else {
		url = ""
	}

	var bowner string
	if len(owner) > 2 {
		bowner = owner[:2] + "/" + owner
	} else {
		bowner = owner + "/" + owner
	}
	path = fmt.Sprintf("%s/packs/%s/%s", ReposDir, bowner, project)
	filename = fmt.Sprintf("pack-%s.%s.pack", owner, project)
	err = nil
	return
}
