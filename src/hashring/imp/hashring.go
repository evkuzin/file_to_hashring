package imp

import (
	"errors"
	"file-to-hashring/src/hashring"
	"file-to-hashring/src/logger"
	"github.com/dgryski/go-jump"
	_ "github.com/lib/pq"
	"hash/fnv"
	"strings"
	"sync"
	"time"
)

const VirtualNodes = 10000

type HashRing struct {
	lock      sync.RWMutex
	servers   []hashring.RingMember
	srv2vNode []int
	hashFunc  func(s string) uint64
	vNodes    int
}

func NewHashRing(servers []hashring.RingMember) hashring.HashRing {
	vNodes := make([]int, VirtualNodes)
	for i := 0; i < VirtualNodes; i++ {
		vNodes[i] = i % len(servers)
	}
	hr := &HashRing{
		servers:   servers,
		vNodes:    VirtualNodes,
		srv2vNode: vNodes,
	}
	return hr
}

func (h *HashRing) GetAllServers() []hashring.RingMember {
	return h.servers
}

func (h *HashRing) GetServer(key string) hashring.RingMember {
	h.lock.RLock()
	defer h.lock.RUnlock()
	return h.servers[h.srv2vNode[h.getVirtualNode(key)]]
}

func (h *HashRing) getVirtualNode(key string) int32 {
	return jump.Hash(h.hash(key), h.vNodes)
}

func (h *HashRing) AddServer(srv hashring.RingMember) error {
	logger.L.Infof("going to rebalance the ring...")
	start := time.Now()
	h.lock.RLock()
	for _, member := range h.servers {
		if strings.EqualFold(member.Name(), srv.Name()) {
			return errors.New("duplicate server")
		}
	}
	h.lock.RLock()
	var keysTotalBefore int
	tmpKeysMap := make(map[int][]string)
	for i, member := range h.servers {
		tmpKeysMap[i] = member.GetAllKeys()
	}
	for _, keys := range tmpKeysMap {
		keysTotalBefore += len(keys)
	}
	h.lock.RUnlock()

	newRing := NewHashRing(append(h.servers, srv))
	logger.L.Infof("new server %s was added", srv.Name())
	vNodes2Reassign := h.vNodes / (len(h.servers) + 1)
	logger.L.Debugf("going to reassign %d vnodes", vNodes2Reassign)
	vNode2KeysMigrate := make(map[int][]string)
	srv2vNodeNew := make([]int, len(h.srv2vNode))
	copy(srv2vNodeNew, h.srv2vNode)
	for i := 0; i < vNodes2Reassign; i++ {
		for vNodeId, srvId := range srv2vNodeNew {
			if i%len(h.servers) == srvId {
				srv2vNodeNew[vNodeId] = len(h.servers)
				vNode2KeysMigrate[vNodeId] = []string{}
				logger.L.Debugf("%d virtual node %d/%d reassigned", i, vNodeId, srvId)
				break
			}
		}
	}

	var keysMigrated int
	for vNode := range vNode2KeysMigrate {
		for _, nodeKey := range h.servers[h.srv2vNode[vNode]].GetAllKeys() {
			if h.getVirtualNode(nodeKey) == int32(vNode) {
				vNode2KeysMigrate[vNode] = append(vNode2KeysMigrate[vNode], nodeKey)
				moveKey(h.servers[h.srv2vNode[vNode]], srv, nodeKey)
				keysMigrated++
			}
		}
	}

	h.lock.RUnlock()
	h.lock.Lock()
	h.dropOldKeys(vNode2KeysMigrate)
	h.servers = newRing.GetAllServers()
	h.srv2vNode = srv2vNodeNew
	h.lock.Unlock()
	h.lock.RLock()
	var keysTotal int
	tmpKeysMap = make(map[int][]string)
	for i, member := range h.servers {
		tmpKeysMap[i] = member.GetAllKeys()
	}
	for _, keys := range tmpKeysMap {
		keysTotal += len(keys)
	}
	h.lock.RUnlock()
	logger.L.Infof(
		"Rebalancing took %s. Total keys before/after: %d/%d. Moved keys: %d",
		time.Now().Sub(start).String(),
		keysTotalBefore,
		keysTotal,
		keysMigrated,
	)
	return nil
}

func moveKey(oldSrv hashring.RingMember, newSrv hashring.RingMember, key string) {
	logger.L.Debugf(
		"moving key %s from %s to %s",
		key,
		oldSrv.Name(),
		newSrv.Name(),
	)

	data, err := oldSrv.GetData(key)
	if err != nil {
		logger.L.Errorf("oops, cant get data from the old server. try to implement retries next time: %s", err)
		return
	}
	err = newSrv.Put(key, data)
	if err != nil {
		logger.L.Errorf("oops, something went wrong: %s", err)
		return
	}
}

func (h *HashRing) VNodes() int {
	return h.vNodes
}

func (h *HashRing) dropOldKeys(keysMap map[int][]string) {
	for vNode, keys := range keysMap {
		for _, key := range keys {
			logger.L.Debugf(
				"dropping key %s from the old keyring",
				key,
			)
			h.servers[h.srv2vNode[vNode]].Delete(key)
		}
	}
}

func (h *HashRing) InitHashFunc(f func(s string) uint64) {
	h.hashFunc = f
}

func (h *HashRing) hash(key string) uint64 {
	if h.hashFunc == nil {
		return h.defaultHashFunc(key)
	} else {
		return h.hashFunc(key)
	}
}

func (h *HashRing) defaultHashFunc(key string) uint64 {
	hash := fnv.New64()
	_, err := hash.Write([]byte(key))
	if err != nil {
		logger.L.Fatalf("oops: %s", err)
	}
	return hash.Sum64()
}
