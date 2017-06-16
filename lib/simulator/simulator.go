package sim

import (
	"github.com/ryankurte/owns/lib/config"
	"github.com/ryankurte/owns/lib/connector"
	"github.com/ryankurte/owns/lib/engine"
	"github.com/ryankurte/owns/lib/runner"
	"log"
)

// Simulator instance
type Simulator struct {
	engine *engine.Engine
	runner *runner.Runner
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

	log.Printf("Starting runnable clients")

	// Launch clients via runner
	err = r.Start()
	if err != nil {
		return nil, err
	}

	log.Printf("Configuring simulation engine")

	// Run engine setup
	err = e.Setup(true)
	if err != nil {
		// Stop runnables
		r.Stop()

		return nil, err
	}

	return &Simulator{e, r}, nil
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

	s.runner.Stop()
}
