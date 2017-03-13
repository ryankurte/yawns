package connectors

import (
	"gopkg.in/zeromq/goczmq.v4"
	"log"
)

// ZMQConnector is a connector instance using ZMQ messaging
type ZMQConnector struct {
	ch *goczmq.Channeler
}

const (
	// DefaultBindAddr default address to bind zmq listener
	DefaultBindAddr string = "tcp://*:6666"
)

// NewZMQConnector creates a new ZMQ based connector instance
func NewZMQConnector(bindAddr string) (*ZMQConnector, error) {

	// Create listener socket
	ch := goczmq.NewRouterChanneler(bindAddr)

	c := ZMQConnector{ch: ch}

	return &c, nil
}

func (c *ZMQConnector) run() {
	for {
		select {
			case p, ok := <- c.ch.RecvChan:
				if ok {
					
				}
		}
	}
}

// Close a ZMQ connector instance
func (c *ZMQConnector) Close() {
	c.sock.Destroy()
}
