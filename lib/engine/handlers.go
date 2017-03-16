package engine

import (
	"log"
)

import (
	"github.com/ryankurte/ons/lib/messages"
)

// HandleConnectorMessage Handle messages sent to the engine from the connector
func (e *Engine) HandleConnectorMessage(message *messages.Message) {

	switch message.GetType() {
	case messages.Connected:
		e.OnConnected(message.GetAddress())

	default:
		log.Printf("Engine.HandleConnectorMessage error: unhandled message type (%s)", message.GetType())
	}

}

// OnConnected called when a node connects
func (e *Engine) OnConnected(address string) {
	n, ok := e.nodes[address]
	if !ok {
		return
	}
	log.Printf("Node %s connected", address)
	n.connected = true

	// TODO: call connected plugins

}
