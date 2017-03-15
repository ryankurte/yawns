package engine

import (
	"fmt"
)

import (
	"github.com/ryankurte/ons/lib/config"
	"github.com/ryankurte/ons/lib/helpers"
)

// Event engine type
// Extends configuration Event model for execution
type Event struct {
	// Base node configuration
	*config.Event

	executed bool
}

// NewEvent creates an engine Event using a provided configuration
func NewEvent(u *config.Event) *Event {
	return &Event{
		Event:    u,
		executed: false,
	}
}

// EventHandler interface implemented by modules that can consume events
type EventHandler interface {
	Event(e *Event) error
}

// HandleSetLocationEvent handles a set location event
func HandleSetLocationEvent(node *Node, data map[string]string) error {
	var err error

	node.Location.Lat, err = helpers.ParseFieldToFloat64("lat", data)
	if err != nil {
		return fmt.Errorf("handleEvent error parsing EventSetLocation %s", err)
	}

	node.Location.Lng, err = helpers.ParseFieldToFloat64("lon", data)
	if err != nil {
		return fmt.Errorf("handleEvent error parsing EventSetLocation %s", err)
	}
	return nil
}
