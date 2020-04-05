package types

type Byter interface {
	ToByte() []byte
	FromByte(v []byte) error
	Key() []byte
	SetKey(k []byte) error
}


type SessionedGeter interface {
	NextGroup(prefixlen int, newitem func() Byter) (items []Byter, err error)
	End()
}

type DBer interface {
	NewHashGeter(sub string) SessionedGeter
}