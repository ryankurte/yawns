package connector

import (
	"gopkg.in/zeromq/goczmq.v4"
)

// ZMQBase object
// Common between server and client instances
type ZMQBase struct {
	ch *goczmq.Channeler
}

// NewZMQBase creates a new ZMQ base instance
func NewZMQBase(ch *goczmq.Channeler) *ZMQBase {

	b := ZMQBase{ch}

	return &b
}

// Send a message without appending any further data
func (c *ZMQBase) Send(data []byte) {
	c.ch.SendChan <- [][]byte{data}
}

// SendWithAddress a message to the provided address using the underlying channel
func (c *ZMQBase) SendWithAddress(address string, data []byte) {
	c.ch.SendChan <- [][]byte{[]byte(address), data}
}

// Exit a ZMQBase instance
func (c *ZMQBase) Exit() {
	c.ch.Destroy()
}
