package connector

import (
	"log"
	"reflect"

	"github.com/zeromq/goczmq"
)

const (
	//DefaultIPCAddress default address for IPC communication
	DefaultIPCAddress = "ipc:///ons"
)

// ZMQConnector is a connector instance using ZMQ messaging
type ZMQConnector struct {
	ch         *goczmq.Channeler
	clients    map[string][]byte
	InputChan  chan interface{}
	OutputChan chan interface{}
}

// NewZMQConnector creates a new ZMQ based connector instance and binds a connector instance and handler to the provided address
func NewZMQConnector(bindAddress string) *ZMQConnector {
	c := ZMQConnector{}

	c.clients = make(map[string][]byte)
	c.InputChan = make(chan interface{}, 1024)
	c.OutputChan = make(chan interface{}, 1024)

	c.ch = goczmq.NewRouterChanneler(bindAddress)

	go c.Run()

	return &c
}

// sendMsg sends an ONS message to the provided client by address
// Note that address lookup is not available until the server has received a message from each client
func (c *ZMQConnector) sendMsg(address string, data []byte) {
	// Lookup ZMQ ID by address
	id, ok := c.clients[address]
	if !ok {
		return
	}

	// Send message via channel
	c.ch.SendChan <- [][]byte{id, data}
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
			err := c.handleIncoming(p)
			if err != nil {
				log.Printf("Parsing error: %s", err)
			}

		// Handle control messages from other components
		case p, ok := <-c.InputChan:
			if !ok {
				log.Printf("channel error")
				break
			}
			//log.Printf("TX from server: %+v", p)
			err := c.handleOutgoing(p)
			if err != nil {
				log.Printf("Parsing error: %s", err)
			}
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
