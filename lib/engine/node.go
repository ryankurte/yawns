package engine

// Location is a world location in floating point degrees with altitude in meters
type Location struct {
	Lat float64
	Lng float64
	Alt float64
}

// Node base type
type Node struct {
	Address    string
	Location   Location
	Executable string
	Arguments  string

	connected bool
	received  uint32
	sent      uint32
}
