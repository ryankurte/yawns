package zmq

import (
	"gopkg.in/zeromq/goczmq.v4"
	"log"
)

// ClientReceiver interface implemented by clients
type ClientReceiver interface {
	Receive(data []byte)
}

// ZMQClient Example Client Implementation
type ZMQClient struct {
	*ZMQBase
	address  string
	receiver ClientReceiver
}

// NewZMQClient creates a new ZMQ client instance
func NewZMQClient(clientAddress, bindAddress string, receiver ClientReceiver) *ZMQClient {
	ch := goczmq.NewDealerChanneler(bindAddress)

	base := NewZMQBase(ch)

	return &ZMQClient{base, clientAddress, receiver}
}

// Send sends a message to the server
func (c *ZMQClient) Send(data []byte) {
	c.SendWithAddress(c.address, data)
}

// Run the ZMQ base
func (c *ZMQClient) Run() {
	for {
		select {
		case p, ok := <-c.ch.RecvChan:
			if !ok {
				log.Printf("channel error")
				break
			}

			log.Printf("Received: %+v", p)
			c.receiver.Receive(p[0])
		}
	}
}
