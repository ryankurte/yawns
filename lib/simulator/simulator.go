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

	// Load connector
	c := connector.NewZMQConnector()

	// Create an underlying engine
	e := engine.NewEngine(c)

	// Initialise connector
	c.Init(o.BindAddr, e)

	// Create client runner
	r := runner.NewRunner()

	// Load configuration file
	config, err := config.LoadConfigFile(o.ConfigFile)
	if err != nil {
		return nil, err
	}

	args := make(map[string]string)
	args["server"] = o.ClientAddr

	// Load configuration into engine
	e.LoadConfig(config)

	// Load configuration into runner
	r.LoadConfig(config, args)

	// Launch runner
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
