package sim

import (
	"log"
	"time"

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

	log.Printf("[INFO] Starting OWNS")

	log.Printf("[DEBUG] Loading configuration file")

	// Load configuration file
	config, err := config.LoadConfigFile(o.ConfigFile)
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] Creating simulation engine")

	// Create the underlying engine
	e := engine.NewEngine(config)

	log.Printf("[DEBUG] Creating connector layer")

	// Load and bind connector
	c := connector.NewZMQConnector(o.BindAddr)
	e.BindConnectorChannels(c.OutputChan, c.InputChan)

	// Add client address to args
	args := make(map[string]string)
	args["server"] = o.ClientAddr

	log.Printf("[DEBUG] Creating client runner")

	// Create and bind client runner
	r := runner.NewRunner(config, config.Defaults.Exec, args)
	e.BindRunnerChannel(r.OutputChan)

	log.Printf("[DEBUG] Initialising plugins")

	// Create plugins
	if c, ok := config.Plugins["pcap"]; ok {
		pcap, err := plugins.NewPCAPPlugin(config.Medium.Bands, time.Now(), c)
		if err != nil {
			return nil, err
		}
		e.BindPlugin(pcap)
	}

	log.Printf("[DEBUG] Launching clients")

	// Launch clients via runner
	err = r.Start()
	if err != nil {
		return nil, err
	}

	log.Printf("[DEBUG] Creating simulation medium")

	m, err := medium.NewMedium(&config.Medium, config.TickRate, &config.Nodes)
	if err != nil {
		return nil, err
	}

	e.BindMedium(m)
	go m.Run()

	log.Printf("[DEBUG] Configuring simulation engine")

	// Run engine setup
	err = e.Setup(true)
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] Setup complete")

	return &Simulator{e, r, m}, nil
}

// Info displays simulation information
func (s *Simulator) Info() {
	s.engine.Info()
	s.runner.Info()
}

// Run launches a simulation
func (s *Simulator) Run() error {
	log.Printf("[INFO] Launching Simulation Instance")

	// Run engine
	err := s.engine.Run()
	if err != nil {
		return err
	}

	log.Printf("[INFO] Simulation complete")
	return nil
}

// Close the simulation
func (s *Simulator) Close() {
	log.Printf("[INFO] Exiting OWNS")

	s.medium.Stop()

	s.runner.Stop()

	s.engine.Close()
}
