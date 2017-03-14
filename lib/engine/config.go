package engine

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"time"
)

// UpdateAction type for valid update actions
type UpdateAction string

const (
	// Set a node location
	UpdateSetLocation UpdateAction = "set-location"
)

// Update struct defines changes to the system
type Update struct {
	// Simulation time at which the update action should be executed
	TimeStamp time.Duration
	// Node address for update to be applied
	Node string
	// Update action to be executed
	Action UpdateAction
	// Update data, parsed based on action
	Data map[string]string
}

// Config Engine configuration
type Config struct {
	// Configuration Name
	Name string
	// End time in ms
	EndTime time.Duration
	// Nodes definitions for the engine
	Nodes []Node
	// Update actions to execute when running
	Updates []Update
}

// LoadConfigFile loads an engine configuration from a config file
func LoadConfigFile(file string) (*Config, error) {

	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Printf("LoadConfig error loading file (%s)", err)
		return nil, err
	}

	c := Config{}

	err = yaml.Unmarshal(data, &c)
	if err != nil {
		log.Printf("LoadConfig error parsing file (%s)", err)
		return nil, err
	}

	return &c, nil
}

// WriteConfigFile writes an engine configuration to a config file
func WriteConfigFile(file string, c *Config) error {

	data, err := yaml.Marshal(c)
	if err != nil {
		log.Printf("LoadConfig error parsing config (%s)", err)
		return err
	}

	err = ioutil.WriteFile(file, data, 0644)
	if err != nil {
		log.Printf("LoadConfig error writing file (%s)", err)
		return err
	}

	return nil
}

// Info prints information about the config to stdout
func (c *Config) Info() {
	log.Printf("Config Name: %s", c.Name)
	log.Printf("  - Nodes: %d", len(c.Nodes))
	log.Printf("  - Updates: %d", len(c.Updates))
}
