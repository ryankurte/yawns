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

	Sent, Received uint32 // Sent and Received packet count
}

type Link struct {
	A, B   int
	Fading float64
}

func GetNodeBounds(nodes []Node) (Location, Location) {
	min, max := nodes[0].Location, nodes[0].Location
	for _, n := range nodes {
		if n.Location.Lat < min.Lat {
			min.Lat = n.Location.Lat
		}
		if n.Location.Lng < min.Lng {
			min.Lng = n.Location.Lng
		}
		if n.Location.Lat > max.Lat {
			max.Lat = n.Location.Lat
		}
		if n.Location.Lng > max.Lng {
			max.Lng = n.Location.Lng
		}
	}
	return min, max
}
