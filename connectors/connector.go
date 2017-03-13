package connectors

import (
	"fmt"
	"github.com/ryankurte/ons/connectors/zmq"
)

// Connector defines the interface that connectors must implement
type Connector interface {
	Init(bindAddress string, h interface{}) error
	Send(address string, data []byte)
}

var connectors map[string]Connector

// Initialise connector module
func init() {
	connectors = make(map[string]Connector)
	connectors["zmq"] = zmq.NewZMQConnector()
}

// GetConnector Fetches a connector by name
func GetConnector(name string) (Connector, error) {
	c, ok := connectors[name]
	if !ok {
		return nil, fmt.Errorf("Connector %s not found", name)
	}
	return c, nil
}
