package engine

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ryankurte/owns/lib/config"
	"github.com/ryankurte/owns/lib/plugins"
)

// Engine is the base simulation engine
type Engine struct {
	nodes map[string]*Node

	Updates []*Update

	medium        Medium
	pluginManager *plugins.PluginManager

	connectorReadCh  chan interface{}
	connectorWriteCh chan interface{}

	runnerLogCh chan string

	startTime   time.Time
	currentTime time.Time
	endTime     time.Duration
	tickRate    time.Duration
}

// NewEngine creates a new engine instance
func NewEngine(c *config.Config) *Engine {
	// Create engine object
	e := Engine{}

	e.loadConfig(c)

	e.pluginManager = plugins.NewPluginManager()

	return &e
}

func (e *Engine) BindMedium(m Medium) {
	e.medium = m
}

func (e *Engine) BindConnectorChannels(read, write chan interface{}) {
	e.connectorReadCh = read
	e.connectorWriteCh = write
}

func (e *Engine) BindRunnerChannel(logCh chan string) {
	e.runnerLogCh = logCh
}

func (e *Engine) BindPlugin(p interface{}) error {
	return e.pluginManager.BindPlugin(p)
}

// LoadConfig Loads a simulation config
func (e *Engine) loadConfig(c *config.Config) {

	// Load settings
	e.tickRate = c.TickRate
	e.endTime = c.EndTime

	// Create map of nodes
	e.nodes = make(map[string]*Node)
	for _, n := range c.Nodes {
		node := Node{
			Node:      &n,
			connected: false,
			received:  0,
			sent:      0,
		}
		e.nodes[n.Address] = &node
	}

	// Create Update array
	e.Updates = make([]*Update, len(c.Updates))
	for i, u := range c.Updates {
		Update := NewUpdate(&u)
		e.Updates[i] = Update
	}

	e.endTime = c.EndTime
}

// Info prints engine information
func (e *Engine) Info() {
	log.Printf("Engine Info")
	log.Printf("  - End Time: %d ms", e.endTime)
	log.Printf("  - Nodes: %d", len(e.nodes))
	log.Printf("  - Updates: %d", len(e.Updates))
}

func (e *Engine) handleUpdate(d time.Duration, addresses []string, action config.UpdateAction, data map[string]string) error {
	for _, address := range addresses {
		err := e.handleNodeUpdate(d, address, action, data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) handleNodeUpdate(d time.Duration, address string, action config.UpdateAction, data map[string]string) error {
	// Fetch matching node
	node, ok := e.nodes[address]
	if !ok {
		return fmt.Errorf("handleUpdate node %s not found", address)
	}

	log.Printf("UPDATE %s %s", address, string(action))

	// Handle actions
	var err error
	switch action {
	case config.UpdateSetLocation:
		err = HandleSetLocationUpdate(node, data)

	default:
		e.pluginManager.OnUpdate(d, action, address, data)
	}

	// Update node instance in storage
	e.nodes[address] = node

	return err
}

func (e *Engine) getNode(address string) (*Node, error) {
	if node, ok := e.nodes[address]; ok {
		return node, nil
	}
	return nil, fmt.Errorf("Node %s not found", address)
}

// Setup engine for simulation
func (e *Engine) Setup(wait bool) error {
	if !wait {
		return nil
	}

	now := time.Second * 0

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	// Await node connections
	log.Printf("[INFO] Setup: Awaiting node connections...")

setup:
	for {
		ready := e.Ready()
		if ready {
			break setup
		}

		select {
		// Loop / recheck
		case <-time.After(1 * time.Second):
			continue

		// Connector inputs
		case message, ok := <-e.connectorReadCh:
			if !ok {
				log.Printf("[ERROR] Connector channel error")
				break setup
			}
			e.medium.Send() <- message
			e.HandleConnectorMessage(now, message)

		// Medium outputs
		case message, ok := <-e.medium.Receive():
			if !ok {
				log.Printf("[ERROR] Connector channel error")
				break setup
			}
			e.connectorWriteCh <- message
			e.HandleMediumMessage(now, message)

		// Runner log inputs
		case line, ok := <-e.runnerLogCh:
			if !ok {
				log.Printf("[ERROR] Runner channel error")
				break setup
			}
			log.Printf("Runner: %s", line)

		// Interrupt channel
		case <-ch:
			return fmt.Errorf("[ERROR] Engine interrupted awaiting node connections")
		// Timeout channel
		case <-time.After(1 * time.Minute):
			return fmt.Errorf("[ERROR] Engine timeout awaiting node connections")
		}
	}

	log.Printf("[INFO] Setup: All nodes connected")

	return nil
}

// Handle Updates at a given tick
func (e *Engine) handleUpdates(d time.Duration) {
	for i, u := range e.Updates {
		// If the time has passed and the Update has not been executed
		if d >= u.TimeStamp && !u.executed {

			log.Printf("[INFO] Executing Update %s (%s)", u.Action, u.Comment)

			// Execute the Update
			err := e.handleUpdate(d, u.Nodes, u.Action, u.Data)
			if err != nil {
				log.Printf("[ERROR] Update error: %s", err)
			}

			// Update the Update list
			u.executed = true
			e.Updates[i] = u
		}
	}
}

// Run the engine
func (e *Engine) Run() error {

	interruptCh := make(chan os.Signal)
	signal.Notify(interruptCh, syscall.SIGINT, syscall.SIGTERM)

	// Run simulation
	e.startTime = time.Now()
	log.Printf("[INFO] Simulation: starting")
	endTimer := time.After(e.endTime)

	runTimer := time.NewTicker(e.tickRate)
	defer runTimer.Stop()

running:
	for {
		select {
		// Exit once endtime has occurred
		case <-endTimer:
			log.Printf("[INFO] Simulation: completed")
			break running

		// Connector inputs
		case message, ok := <-e.connectorReadCh:
			if !ok {
				log.Printf("[ERROR] Connector channel error")
				break running
			}
			e.medium.Send() <- message
			e.HandleConnectorMessage(time.Now().Sub(e.startTime), message)

		case message, ok := <-e.medium.Receive():
			if !ok {
				log.Printf("[ERROR] Medium output channel error")
				break running
			}
			e.connectorWriteCh <- message
			e.HandleMediumMessage(time.Now().Sub(e.startTime), message)

		// Runner log inputs
		case line, ok := <-e.runnerLogCh:
			if !ok {
				log.Printf("[ERROR] Runner channel error")
				break running
			}
			log.Printf("Runner: %s", line)

		// Handle command line interrupts
		case <-interruptCh:
			log.Printf("[INFO] Simulation: interrupted after %s", time.Now().Sub(e.startTime))
			break running

		// Simulation Update ticks
		case t := <-runTimer.C:
			d := t.Sub(e.startTime)
			e.handleUpdates(d)
		}
	}

	return nil
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

func (e *Engine) Close() {
	e.pluginManager.OnClose()
}
