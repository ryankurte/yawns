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

import (
	"github.com/ryankurte/ons/lib/config"
)

// Engine is the base simulation engine
type Engine struct {
	nodes       map[string]Node
	updates     []*Update
	startTime   time.Time
	currentTime time.Time
	endTime     time.Duration
	connector   Connector
}

// NewEngine creates a new engine instance
func NewEngine(c Connector) *Engine {
	// Create engine object
	e := Engine{
		connector: c,
	}

	return &e
}

// LoadConfig Loads a simulation config
func (e *Engine) LoadConfig(c *config.Config) {
	e.nodes = make(map[string]Node)

	// Create map of nodes
	for _, n := range c.Nodes {
		node := Node{
			Node:      &n,
			connected: false,
			received:  0,
			sent:      0,
		}
		e.nodes[n.Address] = node
	}

	// Create update array
	for _, u := range c.Updates {
		update := NewUpdate(&u)
		e.updates = append(e.updates, update)
	}

	e.endTime = c.EndTime
}

// Info prints engine information
func (e *Engine) Info() {
	log.Printf("Engine Info")
	log.Printf("  - End Time: %d ms", e.endTime)
	log.Printf("  - Nodes: %d", len(e.nodes))
	log.Printf("  - Updates: %d", len(e.updates))
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

func (e *Engine) handleUpdate(addresses []string, action config.UpdateAction, data map[string]string) error {
	for _, address := range addresses {
		err := e.handleNodeUpdate(address, action, data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) handleNodeUpdate(address string, action config.UpdateAction, data map[string]string) error {
	// Fetch matching node
	node, ok := e.nodes[address]
	if !ok {
		return fmt.Errorf("handleUpdate node %s not found", address)
	}

	// Handle actions
	var err error
	switch action {
	case config.UpdateSetLocation:
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

// Setup engine for simulation
// This will
func (e *Engine) Setup(wait bool) error {
	if !wait {
		return nil
	}

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

	log.Printf("Connections completed")

	return nil
}

// Handle updates at a given tick
func (e *Engine) handleUpdates(d time.Duration) {
	for i, u := range e.updates {
		// If the time has passed and the update has not been executed
		if d >= u.TimeStamp && !u.executed {

			log.Printf("Executing update %s (%s)", u.Action, u.Comment)

			// Execute the update
			err := e.handleUpdate(u.Nodes, u.Action, u.Data)
			if err != nil {
				log.Printf("Update error: %s", err)
			}

			// Update the update list
			u.executed = true
			e.updates[i] = u
		}
	}
}

// Run the engine
func (e *Engine) Run() error {

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	// Run simulation
	e.startTime = time.Now()
	log.Printf("Starting simulation")

	var lastTime time.Duration
	tickTime := time.Second

running:
	for {
		select {
		// Simulation update ticks
		case <-time.After(lastTime + tickTime):
			lastTime += tickTime
			log.Printf("Simulation tick: %s", lastTime)
			e.handleUpdates(lastTime)

		// Handle command line interrupts
		case <-ch:
			log.Printf("Interrupting simulation after %s", time.Now().Sub(e.startTime))
			break running

		// Exit once endtime has occurred
		case <-time.After(e.endTime):
			log.Printf("Simulation complete")
			break running
		}
	}

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

func (e *Engine) GetCCA(addr string) bool {
	return true
}

// Ready Checks whether the engine is ready to launch
func (e *Engine) Ready() bool {
	ready := true

	// Check that all expected nodes are connected
	for _, n := range e.nodes {
		if !n.connected {
			ready = false
		}
	}

	return ready
}
