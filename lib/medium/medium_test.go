package medium

import (
	"fmt"
	"testing"
	"time"

	"io/ioutil"

	"github.com/ryankurte/ons/lib/config"
	"github.com/ryankurte/ons/lib/messages"
	"github.com/ryankurte/ons/lib/types"

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

	m := NewMedium(&c.Medium, time.Millisecond/10, &c.Nodes)

	go func() {
		m.Run()
	}()

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

	t.Run("Can create transmission instances", func(t *testing.T) {
		msg := messages.Packet{
			Message: messages.Message{Address: "0x0001"},
			RFInfo:  messages.NewRFInfo("Sub1GHz", 1),
			Data:    []byte("test data"),
		}

		now := time.Now()

		band := c.Medium.Bands[msg.Band]
		transmission := NewTransmission(now, &(*m.nodes)[0], &band, msg)

		assert.EqualValues(t, msg.Address, transmission.Origin.Address)
		assert.EqualValues(t, msg.Band, transmission.Band)
		assert.EqualValues(t, msg.Channel, transmission.Channel)
		assert.EqualValues(t, msg.Data, transmission.Data)

		packetTime := time.Duration(float64(len(msg.Data)+int(band.PacketOverhead)) * 8 / float64(band.Baud) * float64(time.Second))
		t.Logf("Packet length %d baud: %s time: %s", len(msg.Data)+int(band.PacketOverhead), band.Baud, packetTime)

		assert.EqualValues(t, now, transmission.StartTime, "Sets start time to now")
		assert.EqualValues(t, packetTime, transmission.PacketTime, "Calculates packet time")
		assert.EqualValues(t, now.Add(packetTime), transmission.EndTime, "Sets end time to now + packet time")
	})

	t.Run("Handles packet transmission", func(t *testing.T) {
		nodeIndex := 0
		node := (*m.nodes)[nodeIndex]
		bandName := "Sub1GHz"

		msg := messages.Packet{
			Message: messages.Message{Address: node.Address},
			RFInfo:  messages.NewRFInfo(bandName, 1),
			Data:    []byte("test data"),
		}

		now := time.Now()
		m.sendPacket(now, msg)

		assert.EqualValues(t, types.TransceiverStateTransmitting, m.transceivers[nodeIndex][bandName], "Sets transceiver state for node")
		assert.EqualValues(t, 1, len(m.transmissions), "Stores transmission instance")

		fmt.Printf("Medium 2: %+v\n", &m)

		transmission := m.transmissions[0]

		assert.EqualValues(t, msg.Address, transmission.Origin.Address)

		now = transmission.EndTime.Add(time.Microsecond)
		m.update(now)

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
