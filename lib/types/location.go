package types

import (
	"fmt"
)

// Location is a world location in floating point degrees with altitude in meters
type Location struct {
	Lat float64 // Latitude
	Lng float64 // Longitude
	Alt float64 // Altitude
}

func (l Location) String() string {
	return fmt.Sprintf("%f,%f,%f", l.Lat, l.Lng, l.Alt)
}
