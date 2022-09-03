package imp

import (
	"file-to-hashring/src/hashring"
	"file-to-hashring/src/logger"
	"github.com/dgryski/go-jump"
	"hash/fnv"
	"sync"
	"time"
)

const ChunksNumber = 10000

type HashRing struct {
	lock     sync.RWMutex
	servers  []hashring.RingMember
	hashFunc func(s string) uint64
	chunks   int
}

func NewHashRing(servers []hashring.RingMember) *HashRing {
	hr := &HashRing{
		servers: servers,
		chunks:  ChunksNumber,
	}
	return hr
}

func (h *HashRing) defaultHashFunc(key string) uint64 {
	hash := fnv.New64()
	_, err := hash.Write([]byte(key))
	if err != nil {
		logger.L.Fatalf("oops: %s", err)
	}
	return hash.Sum64()
}

func (h *HashRing) GetServer(key string) hashring.RingMember {
	h.lock.RLock()
	defer h.lock.RUnlock()
	virtualNode := jump.Hash(h.hash(key), len(h.servers)*h.chunks)
	return h.servers[virtualNode%int32(len(h.servers))]
}

func (h *HashRing) hash(key string) uint64 {
	if h.hashFunc == nil {
		return h.defaultHashFunc(key)
	} else {
		return h.hashFunc(key)
	}
}

func (h *HashRing) addServer(server hashring.RingMember) {
	logger.L.Infof("going to rebalance the ring...")
	start := time.Now()
	h.lock.RLock()
	newRing := NewHashRing(append(h.servers, server))
	logger.L.Infof("new server %s was added", server)
	var keysTotal int
	var keysMigrated int
	tmpKeysMap := make(map[int][]string)
	for i, member := range h.servers {
		tmpKeysMap[i] = member.GetAllKeys()
	}
	tmpMovedKeysMap := make(map[int][]string)
	for srv, keys := range tmpKeysMap {
		tmpMovedKeysMap[srv] = []string{}

		for _, key := range keys {
			keysTotal++
			if h.servers[srv] != newRing.GetServer(key) {
				tmpMovedKeysMap[srv] = append(tmpMovedKeysMap[srv], key)
				moveKey(h, newRing, key)
				keysMigrated++
			}
		}
	}
	h.lock.RUnlock()
	h.lock.Lock()
	h.dropOldKeys(tmpMovedKeysMap)
	h.servers = newRing.servers
	h.lock.Unlock()
	logger.L.Infof(
		"Rebalancing took %s. Total keys: %d. Moved keys: %d",
		time.Now().Sub(start).String(),
		keysTotal,
		keysMigrated,
	)
}

func moveKey(oldRing *HashRing, newRing *HashRing, key string) {
	oldRing.lock.RLock()
	defer oldRing.lock.RUnlock()
	logger.L.Infof(
		"moving key %s from %s to %s",
		key,
		oldRing.GetServer(key).Name(),
		newRing.GetServer(key).Name(),
	)

	data, err := oldRing.GetServer(key).GetData(key)
	if err != nil {
		logger.L.Errorf("oops, cant get data from the old server. try to implement retries next time: %s", err)
		return
	}
	err = newRing.GetServer(key).Put(key, data)
	if err != nil {
		logger.L.Errorf("oops, something went wrong: %s", err)
		return
	}
}

func (h *HashRing) dropOldKeys(keysMap map[int][]string) {
	for srv, keys := range keysMap {
		for _, key := range keys {
			logger.L.Infof(
				"dropping key %s from the old keyring",
				key,
			)
			h.servers[srv].Delete(key)
		}
	}
}

func (h *HashRing) InitHashFunc(f func(s string) uint64) {
	h.hashFunc = f
}

func (h *HashRing) Chunks() int {
	return h.chunks
}
