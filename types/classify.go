package types

type Classifer interface {
	Hit(Hash) bool
	FileNamePrefix() string
}
