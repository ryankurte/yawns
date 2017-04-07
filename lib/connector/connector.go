package connector

import (
	"fmt"
	"gopkg.in/zeromq/goczmq.v4"
	"log"
	"reflect"
)

import (
	"github.com/ryankurte/ons/lib/messages"
)

const (
	DefaultIPCAddress = "ipc:///ons"
)

// ZMQConnector is a connector instance using ZMQ messaging
type ZMQConnector struct {
	ch         *goczmq.Channeler
	clients    map[string][]byte
	InputChan  chan *messages.Message
	OutputChan chan *messages.Message
}

// NewZMQConnector creates a new ZMQ based connector instance and binds a connector instance and handler to the provided address
func NewZMQConnector(bindAddress string) *ZMQConnector {
	c := ZMQConnector{}

	c.clients = make(map[string][]byte)
	c.InputChan = make(chan *messages.Message, 1024)
	c.OutputChan = make(chan *messages.Message, 1024)

	c.ch = goczmq.NewRouterChanneler(bindAddress)

	go c.Run()

	return &c
}

// Send a data message to the provided address
// This wraps SendMsg as a convenience for other modules
func (c *ZMQConnector) Send(address string, data []byte) {
	c.SendMsg(address, onsMessageIDPacket, data)
}

// SendMsg sends an ONS message to the provided client by address
// Note that address lookup is not available until the server has received a message from each client
func (c *ZMQConnector) SendMsg(address string, msgType int, data []byte) {
	// Lookup ZMQ ID by address
	id, ok := c.clients[address]
	if !ok {
		return
	}

	// Create packet type
	// Weirdly these seem to be sent as strings by ZMQ :-/
	t := []byte(fmt.Sprintf("%d", msgType))

	// Send message via channel
	c.ch.SendChan <- [][]byte{id, t, data}
}

// Run the ZMQ connector
func (c *ZMQConnector) Run() {
	for {
		select {
		// Handle protocol messages from clients
		case p, ok := <-c.ch.RecvChan:
			if !ok {
				log.Printf("channel error")
				break
			}

			//log.Printf("RX from client: %+v", p)

			c.handleClientReceive(p)

		// Handle control messages from other components
		case p, ok := <-c.InputChan:
			if !ok {
				log.Printf("channel error")
				break
			}

			//log.Printf("RX from server: %+v", p)

			c.handleMessageReceive(p)
		}
	}
}

// Exit a ZMQConnector instance
func (c *ZMQConnector) Exit() {
	c.ch.Destroy()
}

func (c *ZMQConnector) findClientIDByAddress(address string) []byte {
	id, ok := c.clients[address]
	if !ok {
		return []byte{}
	}
	return id
}

func (c *ZMQConnector) findClientAddressByID(id []byte) string {
	for key, client := range c.clients {
		if reflect.DeepEqual(client, id) {
			return key
		}
	}
	return ""
}
