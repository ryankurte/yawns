package engine

import (
	"github.com/ryankurte/ons/lib/config"
)

// Node base type
type Node struct {
	// Base node configuration
	*config.Node

	// Indicates whether a node has connected to the engine
	connected bool

	// Received packet count
	received uint32
	// Sent packet count
	sent uint32
}

// NewNode creates an engine node using a provided configuration
func NewNode(n *config.Node) *Node {
	return &Node{
		Node: n,
	}
}
