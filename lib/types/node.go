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
	Exec       []string          // Commands to be executed within the node instance

	Sent, Received uint32 // Sent and Received packet count
}

type Nodes []Node

func (n Nodes) FindIndex(address string) (int, bool) {
	for i, v := range n {
		if v.Address == address {
			return i, true
		}
	}
	return 0, false
}

func (n Nodes) Find(address string) (*Node, bool) {
	for _, v := range n {
		if v.Address == address {
			return &v, true
		}
	}
	return nil, false
}

type Link struct {
	A, B   int
	Fading float64
	Meta   interface{}
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

type Links []Link

func (l *Links) Find(a, b int) (float64, bool) {
	for _, v := range *l {
		if v.A == a && v.B == b {
			return v.Fading, true
		}
	}

	return 0, false
}

func (vs Links) Filter(f func(Link) bool) Links {
	vsf := make(Links, 0)
	for _, v := range vs {
		if f(v) {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func (vs Links) Map(f func(Link) Link) Links {
	vsm := make(Links, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}
