package medium

import (
	"testing"
	//"time"

	"io/ioutil"

	"github.com/ryankurte/ons/lib/config"
	//"github.com/ryankurte/ons/lib/messages"
	//"github.com/ryankurte/ons/lib/types"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

func TestMedium(t *testing.T) {

	var c config.Config

	// TODO: test currently depends on the example config file
	// Which might not be ideal, but simple to manage for now
	data, err := ioutil.ReadFile("../../example.yml")
	assert.Nil(t, err)

	err = yaml.Unmarshal(data, &c)
	assert.Nil(t, err)

	m := NewMedium(&c.Medium, &c.Nodes)

	go m.Run()

	t.Run("Maps nodes in config files", func(t *testing.T) {
		if len(*m.nodes) != len(c.Nodes) {
			t.Errorf("Expected 4%d nodes from config file", len(c.Nodes))
		}
	})

	t.Run("Calculates fading between points", func(t *testing.T) {
		fading := m.GetPointToPointFading(c.Medium.Bands["Sub1GHz"], c.Nodes[0], c.Nodes[1])
		assert.InDelta(t, 86, float64(fading), 1.0)

		fading = m.GetPointToPointFading(c.Medium.Bands["Sub1GHz"], c.Nodes[1], c.Nodes[2])
		assert.InDelta(t, 86, float64(fading), 1.0)
	})

	t.Run("Calculates instantaneous connectivity", func(t *testing.T) {
		linkedNodes, _, err := m.GetVisible(c.Nodes[0], c.Medium.Bands["Sub1GHz"])
		assert.Nil(t, err)

		assert.EqualValues(t, 3, len(linkedNodes))
	})

	/*
		t.Run("Sends messages based on link budgets", func(t *testing.T) {

			// Node 1 can communicate with nodes 3 and 5
			message := messages.NewMessage(messages.Packet, "0x0001", []byte("test"))
			m.inCh <- message

			time.Sleep(100 * time.Millisecond)

			// First message will be send complete
			resp := <-m.outCh
			if addr := "0x0001"; resp.GetAddress() != addr {
				t.Errorf("Incorrect message address (actual: %s expected: %s)", resp.GetAddress(), addr)
			}
			if messageType := messages.PacketSent; resp.GetType() != messageType {
				t.Errorf("Incorrect message type (actual: %s expected: %s)", resp.GetType(), messageType)
			}

			// Responses should be in order
			resp = <-m.outCh
			if addr := "0x0002"; resp.GetAddress() != addr {
				t.Errorf("Incorrect message address (actual: %s expected: %s)", resp.GetAddress(), addr)
			}

			resp = <-m.outCh
			if addr := "0x0003"; resp.GetAddress() != addr {
				t.Errorf("Incorrect message address (actual: %s expected: %s)", resp.GetAddress(), addr)
			}

			resp = <-m.outCh
			if addr := "0x0004"; resp.GetAddress() != addr {
				t.Errorf("Incorrect message address (actual: %s expected: %s)", resp.GetAddress(), addr)
			}

		})
	*/
}
