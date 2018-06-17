package plugins

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ryankurte/yawns/lib/config"
)

func TestPCap(t *testing.T) {
	fileName := "./test.pcap"

	c, err := config.LoadConfigFile("../../example.yml")
	assert.Nil(t, err)

	o := make(map[string]interface{})
	o[FileName] = &fileName

	p, err := NewPCAPPlugin(c.Medium.Bands, time.Now(), o)
	if err != nil || p == nil {
		t.Error(err)
		t.FailNow()
	}

	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		assert.Fail(t, "PCap file not created")
	}

	now := time.Second * 1
	err = p.Received(now, "Sub1GHz", "fake-address1", []byte{0xbe, 0xef})
	assert.Nil(t, err, "writing to pcap file failed")

	then := now + time.Millisecond*100
	err = p.Received(then, "IEEE802.15.4-2.4GHz", "fake-address2", []byte{0xca, 0xfe})
	assert.Nil(t, err, "writing to pcap file failed")

	p.Close()
}
