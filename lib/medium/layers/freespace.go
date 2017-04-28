package layers

import (
	"github.com/ryankurte/go-rf"
	"github.com/ryankurte/ons/lib/config"
	"github.com/ryankurte/ons/lib/types"
)

// FreeSpace layer models free space fading at a given frequency
type FreeSpace struct {
}

// NewFreeSpace creates a new free space fading later
func NewFreeSpace() *FreeSpace {
	return &FreeSpace{}
}

// CalculateFading calculates the free space fading for a link
func (fs *FreeSpace) CalculateFading(band config.Band, p1, p2 types.Location) float64 {

	distance := rf.CalculateDistanceLOS(p1.Lat, p1.Lng, p1.Alt, p2.Lat, p2.Lng, p2.Alt)

	return rf.FreeSpaceAttenuationDB(rf.Frequency(band.Frequency), distance)
}
