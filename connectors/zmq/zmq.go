package zmq

import (
	"gopkg.in/zeromq/goczmq.v4"
	"log"
)

// Receiver interface implemented by data receivers
type ServerReceiver interface {
	Receive(address string, data []byte)
}

// ZMQConnector is a connector instance using ZMQ messaging
type ZMQConnector struct {
	*ZMQBase
	clients  map[string][]byte
	receiver ServerReceiver
}

const (
	// DefaultBindAddr default address to bind zmq listener
	DefaultBindAddr string = "tcp://*:6666"
)

// NewZMQConnector creates a new ZMQ based connector instance
func NewZMQConnector(clientAddress, bindAddress string, receiver ServerReceiver) *ZMQConnector {

	ch := goczmq.NewRouterChanneler(bindAddress)

	base := NewZMQBase(ch)

	clients := make(map[string][]byte)

	return &ZMQConnector{base, clients, receiver}
}

// Send sends a message to the provided client by address
// Note that address lookup is not available until the server has received a message from each client
func (c *ZMQConnector) Send(address string, data []byte) {
	id, ok := c.clients[address]
	if !ok {
		return
	}

	c.ZMQBase.ch.SendChan <- [][]byte{id, data}
}

// Run the ZMQ connector
func (c *ZMQConnector) Run() {
	for {
		select {
		case p, ok := <-c.ch.RecvChan:
			if !ok {
				log.Printf("channel error")
				break
			}

			log.Printf("Received: %+v", p)

			id := p[0]
			address := string(p[1])

			c.clients[address] = id

			c.receiver.Receive(address, p[2])
		}
	}
}
