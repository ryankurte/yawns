package engine

import (
	"github.com/ryankurte/ons/lib/config"
)

// Update engine type
// Extends configuration update model for execution
type Update struct {
	// Base node configuration
	*config.Update

	executed bool
}

// NewUpdate creates an engine update using a provided configuration
func NewUpdate(u *config.Update) *Update {
	return &Update{
		Update:   u,
		executed: false,
	}
}
