package engine

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

// Engine is the base simulation engine
type Engine struct {
	nodes       map[string]Node
	updates     []Update
	startTime   time.Time
	currentTime time.Time
	endTime     time.Duration
	connector   Connector
}

// NewEngine creates a new engine instance
func NewEngine() *Engine {
	// Create engine object
	e := Engine{}

	return &e
}

// LoadConfig Loads a simulation config
func (e *Engine) LoadConfig(c *Config) {
	e.nodes = make(map[string]Node)

	// Create map of nodes
	for _, n := range c.Nodes {
		e.nodes[n.Address] = n
	}

	e.updates = c.Updates

	e.endTime = c.EndTime
}

// LoadConfigFile Loads a simulation config from a file
func (e *Engine) LoadConfigFile(fileName string) error {
	c, err := LoadConfigFile(fileName)
	if err != nil {
		return err
	}
	e.LoadConfig(c)
	return nil
}

func (e *Engine) SetConnector(c Connector) {
	e.connector = c
}

// Info prints engine information
func (e *Engine) Info() {
	log.Printf("Engine Info")
	log.Printf("  - End Time: %d ms", e.endTime)
	log.Printf("  - Nodes: %d ms", len(e.nodes))
	log.Printf("  - Updates: %d ms", len(e.updates))
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
func (e *Engine) Run() error {

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	// Await node connections
	log.Printf("Awaiting node connections...")

setup:
	for {
		ready := e.Ready()
		if ready {
			break setup
		}

		select {
		case <-ch:
			return fmt.Errorf("Engine interrupted awaiting node connections")
		case <-time.After(1 * time.Minute):
			return fmt.Errorf("Engine timeout awaiting node connections")
		}
	}

	// Run simulation
	e.startTime = time.Now()
	log.Printf("Starting simulation")

running:
	for {
		// TODO: simulation things here

		// Exit after endtime
		select {
		case <-ch:
			log.Printf("Interrupting simulation after %s", time.Now().Sub(e.startTime))
			break running
		case <-time.After(e.endTime):
			break running
		}
	}

	log.Printf("Exiting simulation")

	return nil
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

// Receive is called when a packet is received
func (e *Engine) Receive(from string, packet []byte) {
	// Check which devices are within range
	for id, _ := range e.nodes {
		if id == from {
			continue
		}

		//hasPath := e.model.HasPath(from, id)
	}
}

// Ready Checks whether the engine is ready to launch
func (e *Engine) Ready() bool {
	ready := true

	// Check that all expected nodes are connected
	for _, n := range e.nodes {
		if n.connected {
			ready = false
		}
	}

	return ready
}
