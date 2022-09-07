package hashring

type RingMember interface {
	Put(name string, raw []byte) error
	GetData(name string) ([]byte, error)
	GetAllKeys() []string
	Delete(key string)
	Name() string
}

type HashRing interface {
	VNodes() int
	AddServer(srv RingMember) error
	GetServer(key string) RingMember
	GetAllServers() []RingMember
}
