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

	return &sim, nil
}

// Info displays simulation information
func (s *Simulator) Info() {
	s.e.Info()
}

// Run launches a simulation
func (s *Simulator) Run() {
	log.Printf("Launching Simulation Instance")

	s.e.Run()

	log.Printf("Simulation complete")
}
