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
	// Simulation time at which the Update action should be executed
	TimeStamp time.Duration
	// Node address for Update to be applied
	Nodes []string
	// Update action to be executed
	Action UpdateAction
	// Update data, parsed based on action
	Data map[string]string
	// Comment for log purposes
	Comment string
}
