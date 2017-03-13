package engine

import (
	"fmt"
	"log"
	"strconv"
)

// Engine is the base simulation engine
type Engine struct {
	nodes   map[string]Node
	updates []Update
}

// NewEngine creates a new engine instance
func NewEngine(c *Config) *Engine {
	nodes := make(map[string]Node)

	// Create map of nodes
	for _, n := range c.Nodes {
		nodes[n.Address] = n
	}

	// Create engine object
	e := Engine{
		nodes:   nodes,
		updates: c.Updates,
	}

	return &e
}

func parseFieldFloat64(name string, data map[string]string) (float64, error) {
	field, ok := data[name]
	if !ok {
		return 0.0, fmt.Errorf("ParseFieldFloat64 error field %s not found", field)
	}

	fieldFloat, err := strconv.ParseFloat(field, 64)
	if err != nil {
		return 0.0, fmt.Errorf("ParseFieldFloat64 error field %s is not a float", field)
	}
	return fieldFloat, nil
}

func (e *Engine) handleUpdate(address string, action UpdateAction, data map[string]string) error {
	// Fetch matching node
	node, ok := e.nodes[address]
	if !ok {
		return fmt.Errorf("handleUpdate node %s not found", address)
	}

	// Handle actions
	var err error
	switch action {
	case UpdateSetLocation:
		node.Location.Lat, err = parseFieldFloat64("lat", data)
		if err != nil {
			return fmt.Errorf("handleUpdate error parsing UpdateSetLocation %s", err)
		}

		node.Location.Lng, err = parseFieldFloat64("lon", data)
		if err != nil {
			return fmt.Errorf("handleUpdate error parsing UpdateSetLocation %s", err)
		}

	default:
		return fmt.Errorf("handleUpdate error, unrecognised action (%s)", action)
	}

	// Update node instance in storage
	e.nodes[address] = node

	return nil
}

func (e *Engine) getNode(address string) (*Node, error) {
	if node, ok := e.nodes[address]; ok {
		return &node, nil
	}
	return nil, fmt.Errorf("Node %s not found", address)
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
