package hashring

type RingMember interface {
	Put(name string, raw []byte) error
	GetSize(name string) (int64, error)
	GetData(name string) ([]byte, error)
	GetAllKeys() []string
	Delete(key string)
	Name() string
}

type HashRing interface {
	Chunks() int
	AddServer(srv RingMember) error
	GetServer(key string) RingMember
	GetAllServers() []RingMember
}
