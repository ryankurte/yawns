/**
 * OpenNetworkSim Medium GeoProvider Package
 * Provides interfaces for a GeoProvider that can be used for medium simulation
 *
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */

package medium

import (
	"github.com/ryankurte/ons/lib/config"
)

// GeoInfo objects returned by GeoProvider.GetLinkInfo
type GeoInfo interface {
	// Check if a connection is line of sight
	IsLineOfSite() bool
	// Fetch the link distance
	GetDistance() float64
}

// GeoProvider must be able to fetch information for a given link
type GeoProvider interface {
	GetLinkInfo(a, b config.Location) (interface{}, error)
}
