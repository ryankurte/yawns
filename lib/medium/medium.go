/**
 * OpenNetworkSim Medium Package
 * Implements wireless medium simulation
 *
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */

package medium

import (
	"github.com/ryankurte/ons/lib/config"
)

// Medium is the wireless medium simulation instance
type Medium struct {
}

// NewMedum creates a new medium instance
func NewMedum() *Medium {
	m := Medium{}

	return &m
}

// GetDistance computes distance between two points using the haversine function
func (m *Medium) GetDistance(a, b *config.Location) {

}
