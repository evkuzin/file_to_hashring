package imp

import (
	"bytes"
	"file-to-hashring/src/hashring"
	"file-to-hashring/src/logger"
	"file-to-hashring/src/storages/inmem"
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
	ring     hashring.HashRing
	testFile []byte
}

type Logger struct {
	t *testing.T
}

func (l Logger) Error(args ...interface{}) {
	l.t.Log(args...)
}

func (l Logger) Errorf(template string, fields ...interface{}) {
	l.t.Logf(template, fields...)
}

func (l Logger) Fatal(args ...interface{}) {
	l.t.Fatal(args...)
}

func (l Logger) Fatalf(template string, fields ...interface{}) {
	l.t.Fatalf(template, fields...)
}

func (l Logger) Debug(args ...interface{}) {
	return
}

func (l Logger) Debugf(template string, fields ...interface{}) {
	return
}

func (l Logger) Info(args ...interface{}) {
	l.t.Log(args...)
}

func (l Logger) Infof(template string, fields ...interface{}) {
	l.t.Logf(template, fields...)
}

func (l Logger) Warn(args ...interface{}) {
	l.t.Log(args...)
}

func (l Logger) Warnf(template string, fields ...interface{}) {
	l.t.Logf(template, fields...)
}

func (h *hashRingTestSuite) SetupSuite() {
	logger.InitLogger(Logger{
		t: h.T(),
	})

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
