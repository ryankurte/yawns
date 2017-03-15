package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"time"
)

// Config Engine configuration
type Config struct {
	// Configuration Name
	Name string
	// End time in ms
	EndTime time.Duration

	// Defaults defines default settings for each node
	Defaults Node

	// Nodes definitions for the engine
	Nodes []Node

	// Update actions to execute when running
	Updates []Update
}

// LoadConfig parses a configuration object and initialises defaults
func loadConfig(c *Config) *Config {

	// Setup node defaults
	for i, n := range c.Nodes {
		if n.Arguments == "" {
			n.Arguments = c.Defaults.Arguments
		}
		if n.Executable == "" {
			n.Executable = c.Defaults.Executable
		}

		c.Nodes[i] = n
	}

	return c
}

// LoadConfigFile loads an engine configuration from a config file
func LoadConfigFile(file string) (*Config, error) {

	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Printf("LoadConfig error loading file (%s)", err)
		return nil, err
	}

	c := &Config{}

	err = yaml.Unmarshal(data, c)
	if err != nil {
		log.Printf("LoadConfig error parsing file (%s)", err)
		return nil, err
	}

	c = loadConfig(c)

	return c, nil
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
