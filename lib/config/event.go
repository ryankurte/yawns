package config

import (
	"time"
)

// EventAction type for valid Event actions
type EventAction string

const (
	// EventSetLocation Set a node location
	EventSetLocation EventAction = "set-location"
)

// Event struct defines changes to the system
type Event struct {
	// Simulation time at which the Event action should be executed
	TimeStamp time.Duration
	// Node address for Event to be applied
	Nodes []string
	// Event action to be executed
	Action EventAction
	// Event data, parsed based on action
	Data map[string]string
	// Comment for log purposes
	Comment string
}
