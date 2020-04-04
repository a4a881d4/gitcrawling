package types

type Byter interface {
	ToByte() []byte
	FromByte(v []byte) error
	Key() []byte
	SetKey(k []byte) error
}
