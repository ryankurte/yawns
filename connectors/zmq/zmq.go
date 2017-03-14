package zmq

import (
	"encoding/binary"
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

	onsMessageRegister uint64 = 1
	onsMessagePacket   uint64 = 2
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

	go c.Run()

	return nil
}

func (c *ZMQConnector) findClientAddressById(id []byte) string {
	for key, client := range c.clients {
		if client == id {
			return key
		}
	}
	return ""
}

// Send sends a message to the provided client by address
// Note that address lookup is not available until the server has received a message from each client
func (c *ZMQConnector) Send(address string, data []byte) {
	// Lookup ZMQ ID by address
	id, ok := c.clients[address]
	if !ok {
		return
	}

	// Create packet type
	t := make([]byte, 4)
	binary.LittleEndian.PutUint32(t, 1)

	c.ZMQBase.ch.SendChan <- [][]byte{id, t, data}
}

func (c *ZMQConnector) receive(data [][]byte) {
	log.Printf("Received: %+v", data)

	if len(data) != 3 {
		log.Printf("Error parsing message, required 3 parts")
		return
	}

	// Fetch ZMQ client ID
	clientID := data[0]

	// Fetch message type
	messageType, err := binary.Uvarint(data[1])
	if err != nil {
		log.Printf("Error parsing message type (%+v)", data[1])
		return
	}

	body := data[2]

	// Handle message type
	switch messageType {
	case onsMessageRegister:
		// Bind address to ID lookup for sending
		address := string(body)
		_, ok := c.clients[address]
		if !ok {
			// Call OnConnect handler if required
			c.handler.OnConnect(address)
			c.clients[address] = clientID
		}
	case onsMessagePacket:
		address := c.findClientAddressById(clientID)
		if address == "" {
			log.Printf("Received message for unknown clientID (%+v)", clientId)
		}
		c.handler.Receive(address, body)

	}

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
