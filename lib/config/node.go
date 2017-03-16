package config

// Location is a world location in floating point degrees with altitude in meters
type Location struct {
	Lat float64
	Lng float64
	Alt float64
}

// Node configuration base type
type Node struct {
	// Address is the node network address
	Address string
	// Location is the physical location of the node
	Location Location

	// Executable is the command to be called by the runner
	Executable string
	// Command is the command to be passed to the executable by the runner (if provided)
	Command string
	// Arguments is a map of the arguments to be provided to the node instance by the runner
	Arguments map[string]string
}
