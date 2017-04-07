package engine

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

import (
	"github.com/ryankurte/ons/lib/config"
	"github.com/ryankurte/ons/lib/messages"
	"github.com/ryankurte/ons/lib/plugins"
)

// Engine is the base simulation engine
type Engine struct {
	nodes         map[string]*Node
	Events        []*Event
	pluginManager *plugins.PluginManager

	connectorReadCh  chan *messages.Message
	connectorWriteCh chan *messages.Message

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

func (e *Engine) BindConnectorChannels(read, write chan *messages.Message) {
	e.connectorReadCh = read
	e.connectorWriteCh = write
}

func (e *Engine) BindRunnerChannel(logCh chan string) {
	e.runnerLogCh = logCh
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

	// Create Event array
	e.Events = make([]*Event, len(c.Events))
	for i, u := range c.Events {
		Event := NewEvent(&u)
		e.Events[i] = Event
	}

	e.endTime = c.EndTime
}

// Info prints engine information
func (e *Engine) Info() {
	log.Printf("Engine Info")
	log.Printf("  - End Time: %d ms", e.endTime)
	log.Printf("  - Nodes: %d", len(e.nodes))
	log.Printf("  - Events: %d", len(e.Events))
}

func (e *Engine) handleEvent(addresses []string, action config.EventAction, data map[string]string) error {
	for _, address := range addresses {
		err := e.handleNodeEvent(address, action, data)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *Engine) handleNodeEvent(address string, action config.EventAction, data map[string]string) error {
	// Fetch matching node
	node, ok := e.nodes[address]
	if !ok {
		return fmt.Errorf("handleEvent node %s not found", address)
	}

	// Handle actions
	var err error
	switch action {
	case config.EventSetLocation:
		err = HandleSetLocationEvent(node, data)

	default:
		return fmt.Errorf("handleEvent error, unrecognised action (%s)", action)
	}

	// Event node instance in storage
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
// TODO: this is a bit broken...
func (e *Engine) Setup(wait bool) error {
	if !wait {
		return nil
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	// Await node connections
	log.Printf("Setup: Awaiting node connections...")

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
				log.Printf("Connector channel error")
				break setup
			}
			e.HandleConnectorMessage(message)

		// Runner log inputs
		case line, ok := <-e.runnerLogCh:
			if !ok {
				log.Printf("Runner channel error")
				break setup
			}
			log.Printf("Runner: %s", line)

		// Interrupt channel
		case <-ch:
			return fmt.Errorf("Engine interrupted awaiting node connections")
		// Timeout channel
		case <-time.After(1 * time.Minute):
			return fmt.Errorf("Engine timeout awaiting node connections")
		}
	}

	log.Printf("Setup: All nodes connected")

	return nil
}

// Handle Events at a given tick
func (e *Engine) handleEvents(d time.Duration) {
	for i, u := range e.Events {
		// If the time has passed and the Event has not been executed
		if d >= u.TimeStamp && !u.executed {

			log.Printf("Executing Event %s (%s)", u.Action, u.Comment)

			// Execute the Event
			err := e.handleEvent(u.Nodes, u.Action, u.Data)
			if err != nil {
				log.Printf("Event error: %s", err)
			}

			// Event the Event list
			u.executed = true
			e.Events[i] = u
		}
	}
}

// Run the engine
func (e *Engine) Run() error {

	interruptCh := make(chan os.Signal)
	signal.Notify(interruptCh, syscall.SIGINT, syscall.SIGTERM)

	// Run simulation
	e.startTime = time.Now()
	log.Printf("Simulation: starting")

	var lastTime time.Duration

running:
	for {
		select {
		// Simulation Event ticks
		case <-time.After(lastTime + e.tickRate):
			lastTime += e.tickRate
			log.Printf("Simulation: tick: %s", lastTime)
			e.handleEvents(lastTime)

		// Connector inputs
		case message, ok := <-e.connectorReadCh:
			if !ok {
				log.Printf("Connector channel error")
				break running
			}
			e.HandleConnectorMessage(message)

		// Runner log inputs
		case line, ok := <-e.runnerLogCh:
			if !ok {
				log.Printf("Runner channel error")
				break running
			}
			log.Printf("Runner: %s", line)

		// Handle command line interrupts
		case <-interruptCh:
			log.Printf("Simulation: interrupted after %s", time.Now().Sub(e.startTime))
			break running

		// Exit once endtime has occurred
		case <-time.After(e.endTime):
			log.Printf("Simulation: completed")
			break running
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
