package imp

import (
	"bytes"
	"file-to-hashring/src/config"
	"file-to-hashring/src/hashring"
	"file-to-hashring/src/logger"
	"file-to-hashring/src/storages/inmem"
	"fmt"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"math/rand"
	"testing"
)

const (
	testFileName      = "test_file"
	testFileChunkSize = 1024
)

type hashRingTestSuite struct {
	suite.Suite
	ring     hashring.HashRing
	testFile []byte
}

func (h *hashRingTestSuite) SetupSuite() {

	zapConf := zap.NewProductionConfig()
	zapConf.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	zapConf.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	logger.InitLogger(&config.Config{Logger: &zapConf})

	h.ring = NewHashRing(inmem.NewHashRingMembersList([]string{"1", "2", "3", "4", "5"}))
	h.testFile = make([]byte, testFileChunkSize*h.ring.VNodes())
	rand.Read(h.testFile)
	r := bytes.NewReader(h.testFile)
	for i := 0; i < h.ring.VNodes(); i++ {
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

	for _, srv := range h.ring.GetAllServers() {
		h.T().Logf("server %s: keys %f%%", srv.Name(), float32(len(srv.GetAllKeys()))/float32(h.ring.VNodes())*100)
	}
}

func TestSampleSuite(t *testing.T) {
	suite.Run(t, new(hashRingTestSuite))
}

func (h *hashRingTestSuite) Test_Ring() {
	testFile := make([]byte, 0)
	for i := 0; i < h.ring.VNodes(); i++ {
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
	err := h.ring.AddServer(inmem.NewInMem("6"))
	if err != nil {
		h.T().Error(err)
	}
	testFile := make([]byte, 0)
	for i := 0; i < h.ring.VNodes(); i++ {
		key := fmt.Sprintf("%s_%d", testFileName, i)

		chunk, err := h.ring.GetServer(key).GetData(key)
		if err != nil {
			h.T().Fatal(err)
		}
		testFile = append(testFile, chunk...)
	}
	if !bytes.Equal(testFile, h.testFile) {
		h.T().Fatal("test files are not equal :( ")
	}
	for _, srv := range h.ring.GetAllServers() {
		h.T().Logf("server %s: keys %f%%", srv.Name(), float32(len(srv.GetAllKeys()))/float32(h.ring.VNodes())*100)
	}
}
