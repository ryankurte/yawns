package layers

import (
	"math/rand"

	"github.com/ryankurte/ons/lib/config"
	"github.com/ryankurte/ons/lib/types"
)

// Random models random fading based on a normal distribution
type Random struct {
}

// NewRandom creates a random fading layer
func NewRandom() *Random {
	return &Random{}
}

// CalculateFading calculates random fading based on an independent normal distribution
func (r *Random) CalculateFading(band config.Band, p1, p2 types.Location) float64 {
	return rand.NormFloat64() * float64(band.RandomDeviation)
}
