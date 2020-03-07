package gitext
import "gopkg.in/src-d/go-git.v4"
func PlainOpen(path string) (*git.Repository, error) {
	return git.PlainOpen(path)
}