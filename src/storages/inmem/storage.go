package inmem

import (
	"errors"
	"file-to-hashring/src/hashring"
)

type inMem struct {
	name string
	kv   map[string][]byte
}

func NewInMem(name string) hashring.RingMember {
	return &inMem{
		name: name,
		kv:   make(map[string][]byte),
	}
}

func NewHashRing(servers []string) []hashring.RingMember {
	hashRingMembers := make([]hashring.RingMember, len(servers))
	for i, server := range servers {
		srv := NewInMem(server)

		hashRingMembers[i] = srv
	}
	return hashRingMembers
}

func (p *inMem) Put(key string, raw []byte) error {
	p.kv[key] = raw
	return nil
}

func (p *inMem) GetData(key string) ([]byte, error) {
	data, ok := p.kv[key]
	if ok {
		return data, nil
	} else {
		return nil, errors.New("key doesnt exist")
	}
}

func (p *inMem) GetSize(key string) (int64, error) {
	data, ok := p.kv[key]
	if ok {
		return int64(len(data)), nil
	} else {
		return 0, errors.New("key doesnt exist")
	}
}

func (p *inMem) GetAllKeys() []string {
	keys := make([]string, len(p.kv))
	var i int
	for key := range p.kv {
		keys[i] = key
		i++
	}
	return keys
}

func (p *inMem) Delete(key string) {
	delete(p.kv, key)
}

func (p *inMem) Name() string {
	return p.name
}
