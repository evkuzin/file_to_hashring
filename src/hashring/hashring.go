package hashring

import (
	"log"
	"net/http"
	"strconv"
	"strings"
)

type HashRing struct {
	servers  []RingMember
	hashFunc func(s string) int
}

func (h *HashRing) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}

type RingMember interface {
	Put(name string, raw []byte) error
	GetSize(name string) (int64, error)
	GetData(name string) ([]byte, error)
}

func New(servers []RingMember) *HashRing {
	hr := &HashRing{
		servers: servers,
	}
	return hr
}

func (h *HashRing) defaultHashFunc(s string) int {
	strSlice := strings.Split(s, "_")
	idxStr := strSlice[len(strSlice)-1]
	idx, err := strconv.Atoi(idxStr)
	if err != nil {
		log.Fatal("wrong name for a file part")
	}
	return idx % len(h.servers)
}

func (h *HashRing) GetServer(key string) RingMember {
	return h.servers[h.hash(key)]
}

func (h *HashRing) hash(key string) int {
	if h.hashFunc == nil {
		return h.defaultHashFunc(key)
	} else {
		return h.hashFunc(key)
	}
}

func (h *HashRing) InitHashFunc(f func(s string) int) {
	h.hashFunc = f
}
