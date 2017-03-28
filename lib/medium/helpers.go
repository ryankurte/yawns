/**
 * OpenNetworkSim Medium Package
 * Implements wireless medium simulation
 *
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */

package medium

import (
	"math/rand"

	"github.com/ryankurte/ons/lib/config"
)

// GetDistance calculates the LoS distance between two locations
// This wraps the RF method with config.Location structures
func GetDistance(a, b *config.Location) float64 {
	return CalculateDistanceLOS(a.Lat, a.Lng, a.Alt, b.Lat, b.Lng, b.Alt)
}

// GetRandomFading Random fading based on a normal distribution with the provided distribution
func GetRandomFading(fadingDist float64) float64 {
	return rand.NormFloat64() * fadingDist
}
