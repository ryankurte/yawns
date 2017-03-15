package config

import (
	"time"
)

// UpdateAction type for valid update actions
type UpdateAction string

const (
	// UpdateSetLocation Set a node location
	UpdateSetLocation UpdateAction = "set-location"
)

// Update struct defines changes to the system
type Update struct {
	// Simulation time at which the update action should be executed
	TimeStamp time.Duration
	// Module to pass update to
	//Module string
	// Node address for update to be applied
	Nodes []string
	// Update action to be executed
	Action UpdateAction
	// Update data, parsed based on action
	Data map[string]string
	// Comment for log purposes
	Comment string
}
