package layers

import (
	"github.com/ryankurte/go-rf"
	"github.com/ryankurte/ons/lib/config"
)

type FreeSpace struct {
}

func NewFreeSpace() *FreeSpace {
	return &FreeSpace{}
}

func (fs *FreeSpace) CalculateFading(freq float64, p1, p2 config.Location) float64 {

	distance := rf.CalculateDistanceLOS(p1.Lat, p1.Lng, p1.Alt, p2.Lat, p2.Lng, p2.Alt)

	return rf.FreeSpaceAttenuationDB(rf.Frequency(freq), distance)
}
