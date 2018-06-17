package engine

import (
	"fmt"
)

import (
	"github.com/ryankurte/yawns/lib/config"
	"github.com/ryankurte/yawns/lib/helpers"
)

// Update engine type
// Extends configuration Update model for execution
type Update struct {
	// Base node configuration
	*config.Update

	executed bool
}

// NewUpdate creates an engine Update using a provided configuration
func NewUpdate(u *config.Update) *Update {
	return &Update{
		Update:   u,
		executed: false,
	}
}

// UpdateHandler interface implemented by modules that can consume Updates
type UpdateHandler interface {
	Update(e *Update) error
}

// HandleSetLocationUpdate handles a set location Update
func HandleSetLocationUpdate(node *Node, data map[string]string) error {
	var err error

	node.Location.Lat, err = helpers.ParseFieldToFloat64("lat", data)
	if err != nil {
		return fmt.Errorf("handleUpdate error parsing UpdateSetLocation %s", err)
	}

	node.Location.Lng, err = helpers.ParseFieldToFloat64("lon", data)
	if err != nil {
		return fmt.Errorf("handleUpdate error parsing UpdateSetLocation %s", err)
	}
	return nil
}
