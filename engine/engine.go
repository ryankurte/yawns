package engine

import (
	"log"
)

// Engine is the base simulation engine
type Engine struct {
	nodes map[string]Node
}

// New creates a new engine instance
func New() *Engine {
	e := Engine{
		make(map[string]Node),
	}
	return &e
}

// Run the engine
func (e *Engine) Run() {
	// Start network interfaces

}

// OnConnect Called when a node connects
func (e *Engine) OnConnect(address string) {
	node, ok := e.nodes[address]
	if !ok {
		log.Printf("Engine.OnConnect: node %s not recognised", address)
		return
	}

	log.Printf("Engine.OnConnect: node %s connected", address)
	node.connected = true
}

// OnDisconnect called when a node disconnects
func (e *Engine) OnDisconnect(address string) {
	node, ok := e.nodes[address]
	if !ok {
		log.Printf("Engine.OnDisconnect: node %s not recognised", address)
		return
	}

	log.Printf("Engine.OnDisconnect: node %s disconnected", address)
	node.connected = false
}

// HandlePacket is called when a packet is received
func (e *Engine) HandlePacket(from string, packet interface{}) {
	// Check which devices are within range
	for id, _ := range e.nodes {
		if id == from {
			continue
		}

		//hasPath := e.model.HasPath(from, id)
	}
}
