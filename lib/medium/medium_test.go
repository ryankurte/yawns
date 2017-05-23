package medium

import (
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

	nodes := make([]*types.Node, len(c.Nodes))
	for i := range c.Nodes {
		nodes[i] = &c.Nodes[i]
	}

	m := NewMedium(&c.Medium, time.Millisecond/10, nodes)

	t.Run("Maps nodes in config files", func(t *testing.T) {
		if len(m.nodes) != len(c.Nodes) {
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
		transmission := NewTransmission(now, m.nodes[0], &band, msg)

		assert.EqualValues(t, msg.Address, transmission.Origin.Address)
		assert.EqualValues(t, msg.Band, transmission.Band)
		assert.EqualValues(t, msg.Channel, transmission.Channel)
		assert.EqualValues(t, msg.Data, transmission.Data)

		packetTime := time.Duration(float64(len(msg.Data)+int(band.PacketOverhead)) * 8 / float64(band.Baud) * float64(time.Second))

		assert.EqualValues(t, now, transmission.StartTime, "Sets start time to now")
		assert.EqualValues(t, packetTime, transmission.PacketTime, "Calculates packet time")
		assert.EqualValues(t, now.Add(packetTime), transmission.EndTime, "Sets end time to now + packet time")
	})

	t.Run("Handles packet transmission", func(t *testing.T) {
		nodeIndex := 0
		node := m.nodes[nodeIndex]
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

		transmission := m.transmissions[0]

		assert.EqualValues(t, msg.Address, transmission.Origin.Address)

		// Cause an update
		now = transmission.EndTime.Add(time.Microsecond)
		m.update(now)

		// Sends SendComplete packet to origin
		select {
		case m := <-m.outCh:
			assert.IsType(t, &messages.SendComplete{}, m)
			resp := m.(*messages.SendComplete)
			assert.EqualValues(t, msg.Address, resp.Address)
			assert.EqualValues(t, msg.Band, resp.Band)

		case <-time.After(time.Second):
			t.Errorf("Timeout waiting for SendComplete output")
		}

		// Distributes messages
		// Node 0 can communicate with nodes 1, 2 and 3
		for i := 1; i < 4; i++ {
			select {
			case o := <-m.outCh:
				assert.IsType(t, &messages.Packet{}, o)
				resp := o.(*messages.Packet)
				assert.EqualValues(t, m.nodes[i].Address, resp.Address)
				assert.EqualValues(t, msg.Band, resp.Band)
				assert.EqualValues(t, msg.Data, resp.Data)

			case <-time.After(time.Second):
				t.Errorf("Timeout waiting for output packet for node %s", m.nodes[i].Address)
			}
		}

		assert.EqualValues(t, 0, len(m.transmissions), "Removes transmission instance")
	})

	t.Run("Handles node movement during packet transmission", func(t *testing.T) {
		nodeIndex := 0
		node := m.nodes[nodeIndex]
		bandName := "Sub1GHz"

		// Shift node 1 out of range
		m.nodes[1].Location.Lat += 1.0

		msg := messages.Packet{
			Message: messages.Message{Address: node.Address},
			RFInfo:  messages.NewRFInfo(bandName, 1),
			Data:    []byte("test data"),
		}

		// Send packet
		now := time.Now()
		m.sendPacket(now, msg)

		packetTime := m.transmissions[0].PacketTime

		// Cause an update
		now = now.Add(packetTime / 2)
		m.update(now)

		assert.EqualValues(t, []bool{false, false, true, true, false}, m.transmissions[0].SendOK)

		// Shift node 1 into range and 2 out of range
		m.nodes[1].Location.Lat -= 1.0
		m.nodes[2].Location.Lat += 1.0

		// Cause another update
		now = now.Add(packetTime / 2)
		m.update(now)

		// Check sendOK flags
		assert.EqualValues(t, []bool{false, false, false, true, false}, m.transmissions[0].SendOK)

		// Cause another update
		now = now.Add(time.Microsecond)
		m.update(now)

		// Sends SendComplete packet to origin
		select {
		case <-m.outCh:
		case <-time.After(time.Second):
			t.Errorf("Timeout waiting for send complete output")
		}

		// Sends message to OK node (3)
		select {
		case o := <-m.outCh:
			assert.IsType(t, &messages.Packet{}, o)
			resp := o.(*messages.Packet)
			assert.EqualValues(t, m.nodes[3].Address, resp.Address)
		case <-time.After(time.Second):
			t.Errorf("Timeout waiting for output packet for node %s", m.nodes[3].Address)
		}
	})

}
