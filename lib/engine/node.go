package engine

import (
	"github.com/ryankurte/yawns/lib/types"
)

// Node base type
type Node struct {
	*types.Node        // Base node configuration
	connected   bool   // Indicates whether a node has connected to the engine
	received    uint32 // Received packet count
	sent        uint32 // Sent packet count
}

// NewNode creates an engine node using a provided configuration
func NewNode(n *types.Node) *Node {
	return &Node{
		Node: n,
	}
}
