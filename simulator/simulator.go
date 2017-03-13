package sim

import (
	"github.com/ryankurte/ons/connectors"
	"github.com/ryankurte/ons/engine"
	"log"
)

// Simulator instance
type Simulator struct {
	e *engine.Engine
}

// NewSimulator creates a simulator instance
func NewSimulator(o *Options) (*Simulator, error) {

	// Create an underlying engine
	e := engine.NewEngine()

	// Load connector
	c, err := connectors.GetConnector(o.Connector)
	if err != nil {
		return nil, err
	}

	c.Init(o.BindAddr, e)
	e.SetConnector(c)

	// Load configuration file
	err = e.LoadConfigFile(o.ConfigFile)
	if err != nil {
		return nil, err
	}

	sim := Simulator{e}

	err = sim.e.Setup(true)
	if err != nil {
		return nil, err
	}

	return &sim, nil
}

// Info displays simulation information
func (s *Simulator) Info() {
	s.e.Info()
}

// Run launches a simulation
func (s *Simulator) Run() error {
	log.Printf("Launching Simulation Instance")

	err := s.e.Run()
	if err != nil {
		return err
	}

	log.Printf("Simulation complete")
	return nil
}
