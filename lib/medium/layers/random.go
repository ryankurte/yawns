package layers

import (
	"math/rand"

	"github.com/ryankurte/ons/lib/config"
)

type Random struct {
	distribution float64
}

func NewRandom(distribution float64) *Random {
	return &Random{distribution}
}

func (r *Random) CalculateFading(freq float64, p1, p2 config.Location) float64 {
	return rand.NormFloat64() * r.distribution
}
