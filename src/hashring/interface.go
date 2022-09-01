package hashring

type RingMember interface {
	Put(name string, raw []byte) error
	GetSize(name string) (int64, error)
	GetData(name string) ([]byte, error)
	GetAllKeys() []string
	Delete(key string)
	Name() string
}
