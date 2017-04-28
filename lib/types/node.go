package types

// Node is a simulated node
type Node struct {
	// Public (loadable) fields
	Address    string            // Address is the node network address
	Location   Location          // Location is the physical location of the node
	Gain       float64           // Gain is the receive and transmit gain modifier in dB (used for different antennas)
	Executable string            // Executable is the command to be called by the runner
	Command    string            // Command is the command to be passed to the executable by the runner (if provided)
	Arguments  map[string]string // Arguments is a map of the arguments to be provided to the node instance by the runner

	sent, received uint32 // Sent and Received packet count
}
