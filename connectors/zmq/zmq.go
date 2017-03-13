package zmq

import (
	"gopkg.in/zeromq/goczmq.v4"
	"log"
)

// Handler interface for server components
type Handler interface {
	OnConnect(address string)
	Receive(address string, data []byte)
}

// ZMQConnector is a connector instance using ZMQ messaging
type ZMQConnector struct {
	*ZMQBase
	clients map[string][]byte
	handler Handler
}

const (
	// DefaultBindAddr default address to bind zmq listener
	DefaultBindAddr string = "tcp://*:6666"
)

// NewZMQConnector creates a new ZMQ based connector instance
func NewZMQConnector() *ZMQConnector {
	c := ZMQConnector{}

	c.clients = make(map[string][]byte)

	return &c
}

// Init binds a connector instance and handler to an address
func (c *ZMQConnector) Init(bindAddress string, h interface{}) error {

	c.handler = h.(Handler)

	ch := goczmq.NewRouterChanneler(bindAddress)

	c.ZMQBase = NewZMQBase(ch)

	return nil
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

func (c *ZMQConnector) receive(data [][]byte) {
	log.Printf("Received: %+v", data)

	// Fetch ZMQ client ID
	id := data[0]

	// Bind address to ID lookup for sending
	address := string(data[1])
	_, ok := c.clients[address]
	if !ok {
		// Call OnConnect handler if required
		c.handler.OnConnect(address)
		c.clients[address] = id
	}

	// Call receive handler
	c.handler.Receive(address, data[2])
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
			c.receive(p)
		}
	}
}
