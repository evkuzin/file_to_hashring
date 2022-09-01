package imp

import (
	"bytes"
	"errors"
	"file-to-hashring/src/hashring"
	"file-to-hashring/src/logger"
	"fmt"
	"github.com/stretchr/testify/suite"
	"math/rand"
	"testing"
)

const (
	testFileName      = "test_file"
	testFileChunkSize = 1024
)

type hashRingTestSuite struct {
	suite.Suite
	ring     *HashRing
	testFile []byte
}

func (h *hashRingTestSuite) SetupSuite() {
	logger.InitLogger()
	h.ring = NewHashRing(newInMemHashRing([]string{"1", "2", "3", "4", "5"}))
	h.testFile = make([]byte, testFileChunkSize*h.ring.chunks)
	rand.Read(h.testFile)
	r := bytes.NewReader(h.testFile)
	for i := 0; i < h.ring.chunks; i++ {
		key := fmt.Sprintf("%s_%d", testFileName, i)
		raw := make([]byte, testFileChunkSize)
		_, err := r.Read(raw)
		if err != nil {
			h.T().Error(err)
		}
		err = h.ring.GetServer(key).Put(key, raw)
		if err != nil {
			h.T().Error(err)
		}
	}

	for _, srv := range h.ring.servers {
		h.T().Logf("server %s: keys %f%%", srv.Name(), float32(len(srv.GetAllKeys()))/float32(h.ring.chunks)*100)
	}
}

func TestSampleSuite(t *testing.T) {
	suite.Run(t, new(hashRingTestSuite))
}

func (h *hashRingTestSuite) Test_Ring() {
	testFile := make([]byte, 0)
	for i := 0; i < h.ring.chunks; i++ {
		key := fmt.Sprintf("%s_%d", testFileName, i)

		chunk, err := h.ring.GetServer(key).GetData(key)
		if err != nil {
			h.T().Error(err)
		}
		testFile = append(testFile, chunk...)
	}
	if !bytes.Equal(testFile, h.testFile) {
		h.T().Errorf("test files are not equal :( ")
	}
}

func (h *hashRingTestSuite) Test_RingRebalance() {
	h.ring.addServer(NewInMem("6"))
	testFile := make([]byte, 0)
	for i := 0; i < h.ring.chunks; i++ {
		key := fmt.Sprintf("%s_%d", testFileName, i)

		chunk, err := h.ring.GetServer(key).GetData(key)
		if err != nil {
			h.T().Error(err)
		}
		testFile = append(testFile, chunk...)
	}
	if !bytes.Equal(testFile, h.testFile) {
		h.T().Errorf("test files are not equal :( ")
	}
	for _, srv := range h.ring.servers {
		h.T().Logf("server %s: keys %f%%", srv.Name(), float32(len(srv.GetAllKeys()))/float32(h.ring.chunks)*100)
	}
}

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

func newInMemHashRing(servers []string) []hashring.RingMember {
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
