package connector

import (
	"fmt"
	"gopkg.in/zeromq/goczmq.v4"
	"log"
	"reflect"
	"strconv"
)

// Handler interface for server components
type Handler interface {
	OnConnect(address string)
	Receive(address string, data []byte)
	GetCCA(address string) bool
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

	// ONSMessageRegister Registration message type, this message binds a socket ID to a client address
	ONSMessageRegister int = 1
	// ONSMessagePacket Packet message type, used for transferring virtual network packets
	ONSMessagePacket int = 2
	// ONSMessageCCAReq CCA request type, requests CCA information from the ONS server
	ONSMessageCCAReq int = 3
	// ONSMessageCCAResp CCA response type, response with CCA info from the ONS server
	ONSMessageCCAResp int = 4
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

func (c *ZMQConnector) findClientAddressByID(id []byte) string {
	for key, client := range c.clients {
		if reflect.DeepEqual(client, id) {
			return key
		}
	}
	return ""
}

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
	c.ZMQBase.ch.SendChan <- [][]byte{id, t, data}
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

func (c *ZMQConnector) receive(data [][]byte) {

	if len(data) != 3 {
		log.Printf("Error parsing message, required 3 parts")
		return
	}

	// Fetch ZMQ client ID
	clientID := data[0]

	// Fetch message type
	messageType, err := strconv.Atoi(string(data[1]))
	if err != nil {
		log.Printf("Error parsing message type (%+v)", data[1])
		return
	}

	// All ONS messages have data[2]
	body := data[2]

	// Handle message type
	switch messageType {
	case ONSMessageRegister:
		// Bind address to ID lookup for sending
		address := string(body)
		_, ok := c.clients[address]
		if !ok {
			// Call OnConnect handler if required
			c.handler.OnConnect(address)
			c.clients[address] = clientID
		}
	case ONSMessagePacket:
		address := c.findClientAddressByID(clientID)
		if address == "" {
			log.Printf("Received message for unknown clientID (%+v)", clientID)
		}
		c.handler.Receive(address, body)

	case ONSMessageCCAReq:
		address := c.findClientAddressByID(clientID)
		if address == "" {
			log.Printf("Received message for unknown clientID (%+v)", clientID)
		}
		cca := c.handler.GetCCA(address)
		dataOut := make([]byte, 1)
		if cca {
			dataOut[0] = 1
		} else {
			dataOut[0] = 0
		}
		c.SendMsg(address, ONSMessageCCAResp, dataOut)

	default:
		log.Printf("Recieved unknown packet type (%d)", messageType)
	}

}
