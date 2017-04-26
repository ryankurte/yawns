package layers

import (
	"github.com/ryankurte/go-rf"
)

type FreeSpace struct {
}

func NewFreeSpace() *FreeSpace {
	return &FreeSpace{}
}

func (fs *FreeSpace) GetFading(freq, lat1, lng1, alt1, lat2, lng2, alt2 float64) float64 {

	distance := rf.CalculateDistanceLOS(lat1, lng1, alt1, lat2, lng2, alt2)

	return rf.FreeSpaceAttenuationDB(rf.Frequency(freq), distance)
}
