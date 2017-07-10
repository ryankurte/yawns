package sim

import (
	"log"

	"github.com/ryankurte/owns/lib/config"
	"github.com/ryankurte/owns/lib/connector"
	"github.com/ryankurte/owns/lib/engine"
	"github.com/ryankurte/owns/lib/medium"
	"github.com/ryankurte/owns/lib/plugins"
	"github.com/ryankurte/owns/lib/runner"
)

// Simulator instance
type Simulator struct {
	engine *engine.Engine
	runner *runner.Runner
	medium *medium.Medium
}

// NewSimulator creates a simulator instance
func NewSimulator(o *Options) (*Simulator, error) {

	// Load configuration file
	config, err := config.LoadConfigFile(o.ConfigFile)
	if err != nil {
		return nil, err
	}

	// Create the underlying engine
	e := engine.NewEngine(config)

	// Load and bind connector
	c := connector.NewZMQConnector(o.BindAddr)
	e.BindConnectorChannels(c.OutputChan, c.InputChan)

	// Add client address to args
	args := make(map[string]string)
	args["server"] = o.ClientAddr

	// Create and bind client runner
	r := runner.NewRunner(config, args)
	e.BindRunnerChannel(r.OutputChan)

	// Create plugins
	if c, ok := config.Plugins["pcap"]; ok {
		pcap, err := plugins.NewPCAPPlugin(c)
		if err != nil {
			return nil, err
		}
		e.BindPlugin(pcap)
	}

	// Launch clients via runner
	err = r.Start()
	if err != nil {
		return nil, err
	}

	log.Printf("Loading medium simulation")
	m := medium.NewMedium(&config.Medium, config.TickRate, &config.Nodes)
	e.BindMedium(m)
	go m.Run()

	log.Printf("Configuring simulation engine")

	// Run engine setup
	err = e.Setup(true)
	if err != nil {
		return nil, err
	}

	log.Printf("Starting runnable clients")

	return &Simulator{e, r, m}, nil
}

// Info displays simulation information
func (s *Simulator) Info() {
	s.engine.Info()
	s.runner.Info()
}

// Run launches a simulation
func (s *Simulator) Run() error {
	log.Printf("Launching Simulation Instance")

	// Run engine
	err := s.engine.Run()
	if err != nil {
		return err
	}

	log.Printf("Simulation complete")
	return nil
}

// Close the simulation
func (s *Simulator) Close() {

	s.medium.Stop()

	s.runner.Stop()

	s.engine.Close()
}
