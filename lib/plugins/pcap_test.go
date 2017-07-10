package plugins

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestPCap(t *testing.T) {
	fileName := "./test.pcap"

	o := make(map[string]interface{})
	o[FileName] = &fileName

	p, err := NewPCAPPlugin(o)
	if err != nil || p == nil {
		t.Error(err)
		t.FailNow()
	}

	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		assert.Fail(t, "PCap file not created")
	}

	now := time.Now()
	err = p.Received(now, "fake-address1", []byte{0xbe, 0xef})
	assert.Nil(t, err, "writing to pcap file failed")

	then := now.Add(time.Second + time.Millisecond)
	err = p.Received(then, "fake-address2", []byte{0xca, 0xfe})
	assert.Nil(t, err, "writing to pcap file failed")

	p.Close()
}
