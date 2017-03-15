package connector

import (
	"fmt"
	"gopkg.in/zeromq/goczmq.v4"
	"log"
	"reflect"
)

const (
	// DefaultBindAddr default address to bind zmq listener
	DefaultBindAddr string = "tcp://*:6666"
)

// Handler interface for server components
type Handler interface {
	OnConnect(address string)
	Receive(address string, data []byte)
	GetCCA(address string) bool
}

// ZMQConnector is a connector instance using ZMQ messaging
type ZMQConnector struct {
	ch      *goczmq.Channeler
	clients map[string][]byte
	handler Handler
	in      chan interface{}
	out     chan interface{}
}

// NewZMQConnector creates a new ZMQ based connector instance
func NewZMQConnector() *ZMQConnector {
	c := ZMQConnector{}

	c.clients = make(map[string][]byte)

	return &c
}

// Init binds a connector instance and handler to an address
func (c *ZMQConnector) Init(bindAddress string, h interface{}) error {

	c.handler = h.(Handler)

	c.ch = goczmq.NewRouterChanneler(bindAddress)

	go c.Run()

	return nil
}

func (c *ZMQConnector) findClientAddressByID(id []byte) string {
	for key, client := range c.clients {
		if reflect.DeepEqual(client, id) {
			return key
		}
	}
	return ""
}

// Send a data message to the provided address
// This wraps SendMsg as a convenience for other modules
func (c *ZMQConnector) Send(address string, data []byte) {
	c.SendMsg(address, ONSMessagePacket, data)
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
		case p, ok := <-c.ch.RecvChan:
			if !ok {
				log.Printf("channel error")
				break
			}
			c.receive(p)
		}
	}
}

// Exit a ZMQConnector instance
func (c *ZMQConnector) Exit() {
	c.ch.Destroy()
}
