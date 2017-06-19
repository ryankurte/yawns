package engine

import (
	"log"
	"reflect"
)

import (
	"github.com/ryankurte/owns/lib/messages"
)

// HandleConnectorMessage Handle messages sent to the engine from the connector
func (e *Engine) HandleConnectorMessage(message interface{}) {

	switch m := message.(type) {
	case *messages.Register:
		e.OnConnected(m.GetAddress())

	case *messages.Packet:
		e.OnReceived(m.GetAddress(), m.Data)

	default:
		log.Printf("Engine.HandleConnectorMessage error: unhandled message type (%s)", reflect.TypeOf(message))
	}
}

// OnConnected called when a node connects
func (e *Engine) OnConnected(address string) {
	node, ok := e.nodes[address]
	if !ok {
		log.Printf("Node registration not found")
		return
	}

	node.connected = true

	// Call connected plugins
	e.pluginManager.OnConnected(address)

}

// OnReceived called when a packet is received from the connector
func (e *Engine) OnReceived(address string, data []byte) {

	// Update stats
	node, ok := e.nodes[address]
	if !ok {
		return
	}

	node.sent++

	// Call plugin
	e.pluginManager.OnReceived(address, data)
}

// OnSend called when a packet is sent to the connector
func (e *Engine) OnSend(address string, data []byte) {
	log.Printf("OnSend called")

	// Update stats
	node, ok := e.nodes[address]
	if !ok {
		return
	}

	node.received++

	e.pluginManager.OnSend(address, data)
}
