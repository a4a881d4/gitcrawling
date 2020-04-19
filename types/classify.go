package types

type Classifer interface {
	Hit(Hash) bool
	NamePrefix() string
}
