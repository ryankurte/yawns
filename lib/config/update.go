package config

import (
	"time"
)

// UpdateAction type for valid Update actions
type UpdateAction string

const (
	// UpdateSetLocation Set a node location
	UpdateSetLocation UpdateAction = "set-location"
	// UpdateSetState sets a state field for a given node and key
	UpdateSetState UpdateAction = "set-state"
	// UpdateCheckState checks a state field for a given node and key
	UpdateCheckState UpdateAction = "check-state"
)

// Update struct defines changes to the system
type Update struct {
	Action    UpdateAction      `yaml:"action"`    // Update action to be executed
	TimeStamp time.Duration     `yaml:"timestamp"` // Simulation time at which the Update action should be executed
	Nodes     []string          `yaml:"nodes"`     // Node address for Update to be applied
	Data      map[string]string `yaml:"data"`      // Update data, parsed based on action
	Comment   string            `yaml:"comment"`   // Comment for log purposes
}
