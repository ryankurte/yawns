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

	case messages.Packet:
		e.OnReceived(message.GetAddress(), message.GetData())

	default:
		log.Printf("Engine.HandleConnectorMessage error: unhandled message type (%s)", message.GetType())
	}

}

// OnConnected called when a node connects
func (e *Engine) OnConnected(address string) {
	log.Printf("OnConnected called")
	node, ok := e.nodes[address]
	if !ok {
		log.Printf("Node registration not found")
		return
	}
	log.Printf("Node %s connected", address)
	node.connected = true
	e.nodes[address] = node

	// Call connected plugins
	e.pluginManager.OnConnected(address)

}

// OnReceived called when a packet is received from the connector
func (e *Engine) OnReceived(address string, data []byte) {
	log.Printf("OnReceived called")

	// Update stats
	node, ok := e.nodes[address]
	if !ok {
		return
	}

	node.sent++
	e.nodes[address] = node

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
	e.nodes[address] = node

	e.pluginManager.OnSend(address, data)
}
