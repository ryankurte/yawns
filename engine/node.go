package engine

import (
	"github.com/golang/geo/s2"
)

// Node base type
type Node struct {
	Address  string
	Location s2.LatLng

	connected bool
	received  uint32
	sent      uint32
}
