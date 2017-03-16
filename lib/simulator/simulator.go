package sim

import (
	"github.com/ryankurte/ons/lib/config"
	"github.com/ryankurte/ons/lib/connector"
	"github.com/ryankurte/ons/lib/engine"
	"github.com/ryankurte/ons/lib/runner"
	"log"
	"time"
)

// Simulator instance
type Simulator struct {
	engine *engine.Engine
	runner *runner.Runner
}

// NewSimulator creates a simulator instance
func NewSimulator(o *Options) (*Simulator, error) {

	// Create the underlying engine
	e := engine.NewEngine()

	// Load and bind connector
	c := connector.NewZMQConnector(o.BindAddr)
	e.BindConnectorChannels(c.OutputChan, c.InputChan)

	// Create and bind client runner
	r := runner.NewRunner()
	e.BindRunnerChannel(r.OutputChan)

	// Load configuration file
	config, err := config.LoadConfigFile(o.ConfigFile)
	if err != nil {
		return nil, err
	}

	// Add client address to args
	args := make(map[string]string)
	args["server"] = o.ClientAddr

	// Load configuration into engine
	e.LoadConfig(config)

	// Load configuration into runner
	r.LoadConfig(config, args)

	// Launch clients via runner
	err = r.Start()
	if err != nil {
		return nil, err
	}

	time.Sleep(100 * time.Millisecond)

	// Run engine setup
	err = e.Setup(false)
	if err != nil {
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
