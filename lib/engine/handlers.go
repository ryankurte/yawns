package engine

import (
	"log"
	"time"

	"github.com/ryankurte/owns/lib/config"
	"github.com/ryankurte/owns/lib/messages"
)

// HandleConnectorMessage Handle messages sent to the engine from the connector
func (e *Engine) HandleConnectorMessage(d time.Duration, message interface{}) {
	// Route messages to the appropriate handler
	switch m := message.(type) {
	case messages.Register:
		e.OnConnected(d, m.GetAddress())
	case messages.Packet:
		e.OnReceived(d, m.Band, m.GetAddress(), m.Data)
	case messages.Event:
		e.OnEvent(d, m.Address, m.Data)
	default:
		e.OnMessage(d, message)
	}
}

// HandleMediumMessage handles messages from the medium emulation module
func (e *Engine) HandleMediumMessage(d time.Duration, message interface{}) {
	switch m := message.(type) {
	case messages.Packet:
		e.OnSend(d, m.Band, m.GetAddress(), m.Data)
	}
}

// OnConnected called when a node connects
func (e *Engine) OnConnected(d time.Duration, address string) {
	node, ok := e.nodes[address]
	if !ok {
		log.Printf("Node registration not found")
		return
	}

	// Set connected state
	node.connected = true

	// Call connected plugins
	e.pluginManager.OnConnected(d, address)
}

// OnReceived called when a packet is received from the connector
func (e *Engine) OnReceived(d time.Duration, band, address string, data []byte) {
	// Update stats
	node, ok := e.nodes[address]
	if !ok {
		return
	}
	node.sent++

	// Call plugins
	e.pluginManager.OnReceived(d, band, address, data)
}

// OnSend called when a packet is sent to the connector
func (e *Engine) OnSend(d time.Duration, band, address string, data []byte) {
	// Update stats
	node, ok := e.nodes[address]
	if !ok {
		return
	}
	node.received++

	// Call plugins
	e.pluginManager.OnSend(d, band, address, data)
}

// OnEvent called for nde events
func (e *Engine) OnEvent(d time.Duration, address string, data string) {
	e.pluginManager.OnEvent(d, address, data)
}

// OnUpdate called for simulation updates
func (e *Engine) OnUpdate(d time.Duration, eventType config.UpdateAction, address string, data map[string]string) {
	e.pluginManager.OnUpdate(d, eventType, address, data)
}

// OnMessage called for all unhandled messages
func (e *Engine) OnMessage(d time.Duration, message interface{}) {
	e.pluginManager.OnMessage(d, message)
}
