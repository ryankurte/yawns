package medium

import (
	"testing"
	"time"

	"github.com/ryankurte/owns/lib/config"
	//"github.com/ryankurte/owns/lib/medium/layers"
	"github.com/ryankurte/owns/lib/messages"
	"github.com/ryankurte/owns/lib/types"

	"github.com/stretchr/testify/assert"
)

func TestMedium(t *testing.T) {

	// TODO: test currently depends on the example config file
	// Which might not be ideal, but simple to manage for now
	c, err := config.LoadConfigFile("./test.yml")
	assert.Nil(t, err)

	nodes := make([]types.Node, len(c.Nodes))
	for i := range c.Nodes {
		nodes[i] = c.Nodes[i]
	}

	m, err := NewMedium(&c.Medium, time.Millisecond/10, &nodes)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	now := time.Now()
	for i := range nodes {
		for b := range c.Medium.Bands {
			m.SetTransceiverState(now, i, b, types.TransceiverStateReceive)
		}
	}

	t.Run("Maps nodes in config files", func(t *testing.T) {
		if len(*m.nodes) != len(c.Nodes) {
			t.Errorf("Expected 4%d nodes from config file", len(c.Nodes))
		}
	})

	t.Run("Calculates fading between points", func(t *testing.T) {
		fading := m.GetPointToPointFading(c.Medium.Bands["Sub1GHz"], c.Nodes[0], c.Nodes[1])
		assert.InDelta(t, 86, float64(fading), 1.0)

		fading = m.GetPointToPointFading(c.Medium.Bands["Sub1GHz"], c.Nodes[1], c.Nodes[2])
		assert.InDelta(t, 87, float64(fading), 1.0)
	})

	t.Run("Can create transmission instances", func(t *testing.T) {
		msg := messages.Packet{
			BaseMessage: messages.BaseMessage{Address: "0x0001"},
			RFInfo:      messages.NewRFInfo("Sub1GHz", 1),
			Data:        []byte("test data"),
		}

		now := time.Now()

		band := c.Medium.Bands[msg.Band]
		transmission := NewTransmission(now, &nodes[0], &band, msg)

		assert.EqualValues(t, msg.Address, transmission.Origin.Address)
		assert.EqualValues(t, msg.Band, transmission.Band)
		assert.EqualValues(t, msg.Channel, transmission.Channel)
		assert.EqualValues(t, msg.Data, transmission.Data)

		packetTime := time.Duration(float64(len(msg.Data)+int(band.PacketOverhead)) * 8 / float64(band.Baud) * float64(time.Second))

		assert.EqualValues(t, now, transmission.StartTime, "Sets start time to now")
		assert.EqualValues(t, packetTime, transmission.PacketTime, "Calculates packet time")
		assert.EqualValues(t, now.Add(packetTime), transmission.EndTime, "Sets end time to now + packet time")
	})

	nodeIndex := 0
	node := nodes[nodeIndex]
	bandName := "Sub1GHz"

	t.Run("Handles packet transmission", func(t *testing.T) {

		msg := messages.Packet{
			BaseMessage: messages.BaseMessage{Address: node.Address},
			RFInfo:      messages.NewRFInfo(bandName, 1),
			Data:        []byte("test data"),
		}

		now := time.Now()
		m.sendPacket(now, msg)

		assert.EqualValues(t, types.TransceiverStateTransmitting, m.transceivers[nodeIndex][bandName].State, "Sets transceiver state for node")
		assert.EqualValues(t, 1, len(m.transmissions), "Stores transmission instance")

		transmission := m.transmissions[0]

		assert.EqualValues(t, msg.Address, transmission.Origin.Address)

		// Cause an update
		now = transmission.EndTime.Add(time.Microsecond)
		m.update(now)

		// Sends SendComplete packet to origin
		CheckSendComplete(t, msg.Address, msg.RFInfo, m.outCh)

		// Distributes messages
		// Node 0 can communicate with nodes 1, 2 and 3
		CheckPacketForward(t, nodes[1].Address, msg.Data, msg.RFInfo, m.outCh)
		CheckPacketForward(t, nodes[3].Address, msg.Data, msg.RFInfo, m.outCh)

		assert.EqualValues(t, 0, len(m.transmissions), "Removes transmission instance")

		assert.EqualValues(t, types.TransceiverStateReceive, m.transceivers[nodeIndex][bandName].State, "Resets transceiver state for node")
	})

	t.Run("Handles node movement during packet transmission", func(t *testing.T) {

		// Shift node 1 out of range
		nodes[1].Location.Lat += 1.0

		msg := messages.Packet{
			BaseMessage: messages.BaseMessage{Address: node.Address},
			RFInfo:      messages.NewRFInfo(bandName, 1),
			Data:        []byte("test data"),
		}

		// Send packet
		now := time.Now()
		m.sendPacket(now, msg)

		packetTime := m.transmissions[0].PacketTime

		// Cause an update
		now = now.Add(packetTime / 2)
		m.update(now)

		assert.EqualValues(t, []bool{false, false, false, true, false, false}, m.transmissions[0].SendOK)

		// Shift node 1 into range and 2 out of range
		nodes[1].Location.Lat -= 1.0
		nodes[2].Location.Lat += 1.0

		// Cause another update
		now = now.Add(packetTime / 2)
		m.update(now)

		// Check sendOK flags
		assert.EqualValues(t, []bool{false, false, false, true, false, false}, m.transmissions[0].SendOK)

		// Cause another update
		now = now.Add(time.Microsecond)
		m.update(now)

		// Sends SendComplete packet to origin
		CheckSendComplete(t, msg.Address, msg.RFInfo, m.outCh)

		// Sends message to OK node (3)
		CheckPacketForward(t, nodes[3].Address, msg.Data, msg.RFInfo, m.outCh)

		// Reset node 2 location
		nodes[2].Location.Lat -= 1.0

		assert.EqualValues(t, 0, len(m.transmissions), "Removes transmission instance")
	})

	t.Run("Handles packet collisions", func(t *testing.T) {

		// Create two packets
		msg1 := messages.Packet{
			BaseMessage: messages.BaseMessage{Address: nodes[1].Address},
			RFInfo:      messages.NewRFInfo(bandName, 1),
			Data:        []byte("test data 1"),
		}
		msg2 := messages.Packet{
			BaseMessage: messages.BaseMessage{Address: nodes[2].Address},
			RFInfo:      messages.NewRFInfo(bandName, 1),
			Data:        []byte("test data 2"),
		}

		// Send packets
		now := time.Now()
		m.sendPacket(now, msg1)
		m.sendPacket(now, msg2)

		packetTime := m.transmissions[0].PacketTime

		// Update a couple of times to calculate links
		now = now.Add(packetTime / 2)
		m.update(now)
		now = now.Add(packetTime / 2)
		m.update(now)

		// At this instant collisions have been detected and transmissions not yet removed
		assert.EqualValues(t, []bool{false, false, true, false, false, false}, m.transmissions[0].SendOK)
		assert.EqualValues(t, []bool{false, true, false, false, false, false}, m.transmissions[1].SendOK)

		// Next instant causes transmission to be finalised
		now = now.Add(time.Microsecond)
		m.update(now)

		// Sends SendComplete packet to msg1 origin
		CheckSendComplete(t, msg1.Address, msg1.RFInfo, m.outCh)
		CheckSendComplete(t, msg2.Address, msg2.RFInfo, m.outCh)

		// Clears transmissions
		assert.EqualValues(t, 0, len(m.transmissions), "Removes transmission instances")
	})

	t.Run("Only sends to receiving nodes", func(t *testing.T) {

	})

}

func CheckSendComplete(t assert.TestingT, address string, rfInfo messages.RFInfo, ch chan interface{}, msgAndArgs ...interface{}) {
	sendComplete := messages.NewSendComplete(address, rfInfo.Band, rfInfo.Channel)
	resp := ChannelGet(t, ch, time.Millisecond, msgAndArgs...)
	assert.IsType(t, &messages.SendComplete{}, resp, msgAndArgs...)
	assert.EqualValues(t, sendComplete, resp, msgAndArgs...)
}

func CheckPacketForward(t assert.TestingT, address string, data []byte, rfInfo messages.RFInfo, ch chan interface{}, msgAndArgs ...interface{}) {
	forwardedPacket := messages.NewPacket(address, data, rfInfo)
	resp := ChannelGet(t, ch, time.Millisecond, msgAndArgs...)
	assert.IsType(t, &messages.Packet{}, resp, msgAndArgs...)
	if respPacket, ok := resp.(*messages.Packet); ok {
		forwardedPacket.RSSI = respPacket.RSSI
	}
	assert.EqualValues(t, forwardedPacket, resp, msgAndArgs...)
}

func ChannelGet(t assert.TestingT, ch chan interface{}, timeout time.Duration, msgAndArgs ...interface{}) interface{} {
	select {
	case o, ok := <-ch:
		if !ok {
			assert.Fail(t, "Channel closed", msgAndArgs...)
			return nil
		}
		return o

	case <-time.After(time.Second):
		assert.Fail(t, "Channel timeout", msgAndArgs...)
	}
	return nil
}
