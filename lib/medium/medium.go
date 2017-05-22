/**
 * OpenNetworkSim Medium Package
 * Implements wireless medium simulation
 *
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */

package medium

import (
	"fmt"
	"log"
	//"time"

	"github.com/ryankurte/ons/lib/config"
	"github.com/ryankurte/ons/lib/medium/layers"
	"github.com/ryankurte/ons/lib/messages"
	"github.com/ryankurte/ons/lib/types"
)

// Link encapsulates a link between two nodes
type Link struct {
	From, To string
	Distance float64
	Fading   float64
}

// TransceiverState is the state of a virtual transceiver
type TransceiverState string

// Allowed transceiver states
const (
	TransceiverStateIdle         TransceiverState = "idle"
	TransceiverStateReceive      TransceiverState = "receive"
	TransceiverStateReceiving    TransceiverState = "receiving"
	TransceiverStateTransmitting TransceiverState = "transmitting"
)

// BandInfo is rf band information for a given node
type BandInfo struct {
	Name  string
	State TransceiverState
	Gain  float64
}

// Medium is the wireless medium simulation instance
type Medium struct {
	config *config.Medium
	nodes  *[]types.Node
	bands  [][]BandInfo

	layerManager *layers.LayerManager

	inCh  chan *messages.Message
	outCh chan *messages.Message
}

// NewMedium creates a new medium instance
func NewMedium(c *config.Medium, nodes *[]types.Node) *Medium {
	// Create base medium object
	m := Medium{
		config:       c,
		inCh:         make(chan *messages.Message, 128),
		outCh:        make(chan *messages.Message, 128),
		bands:        make([][]BandInfo, len(*nodes)),
		layerManager: layers.NewLayerManager(),
		nodes:        nodes,
	}

	// Create BandInfo for each node and band
	for i := 0; i < len(*nodes); i++ {
		m.bands[i] = make([]BandInfo, len(c.Bands))
		j := 0
		for k := range c.Bands {
			m.bands[i][j] = BandInfo{
				Name:  k,
				State: TransceiverStateIdle,
			}
			j++
		}
	}

	// Load medium simulation layers
	m.layerManager.BindLayer(layers.NewFreeSpace())
	m.layerManager.BindLayer(layers.NewRandom())

	return &m
}

// GetPointToPointFading fetches the (instantaneous) fading between two nodes at a given frequency
func (m *Medium) GetPointToPointFading(band config.Band, n1, n2 types.Node) types.Attenuation {
	return types.Attenuation(m.layerManager.CalculateFading(band, n1.Location, n2.Location))
}

// GetVisible fetches a list of visible nodes for a provided source node
func (m *Medium) GetVisible(source types.Node, band config.Band) ([]*types.Node, []float64, error) {
	visible := make([]*types.Node, 0)
	attenuation := make([]float64, 0)

	// Iterate through node array
	for _, node := range *m.nodes {
		// Skip source node
		if node.Address == source.Address {
			continue
		}

		// Calculate fading and add links where appropriate
		fading := m.GetPointToPointFading(band, source, node)
		if band.LinkBudget > fading {
			visible = append(visible, &node)
			attenuation = append(attenuation, float64(fading))
		}
	}

	return visible, attenuation, nil
}

func (m *Medium) Run() {
	log.Printf("Medium running")
running:
	for {
		select {
		case message, ok := <-m.inCh:
			if !ok {
				log.Printf("Medium input channel closed")
				break running
			}

			m.handleMessage(message)
		}
	}
	log.Printf("Medium exited")
}

func (m *Medium) handleMessage(message interface{}) {
	switch m := message.(type) {
	case messages.Packet:
		log.Printf("Received packet: %+v", m)
		//m.sendPacket(message.GetAddress(), message.GetData())
	default:
		log.Printf("Medium unhandled message: %+v", message)
	}
}

func (m *Medium) sendPacket(fromAddr string, bandName string, data []byte) error {

	// Locate source node and source band info
	source, sourceInfo, err := m.getNodeAndBandInfo(fromAddr, bandName)
	if err != nil {
		return err
	}

	// Locate matching band
	band, ok := m.config.Bands[bandName]
	if !ok {
		return fmt.Errorf("Medium error: no matching band configured (%s)", bandName)
	}

	// Set transmitting state
	sourceInfo.State = TransceiverStateTransmitting

	// Set timeout for packet sent response
	//packetTime := float64((len(data) + band.PacketOverhead)) / float64(band.Baud)
	source.Transmitting = true

	// Build a list of viable links
	links, _, err := m.GetVisible(*source, band)

	fmt.Printf("Viable links: %+v", links)
	/*
		// Run callback after packet send has completed
		time.AfterFunc(time.Duration(packetTime)*time.Second, func() {
			// Send packet-sent message to application
			source.Transmitting = false
			m.outCh <- messages.Packet(messages.PacketSent, source.Address, []byte{})

			// Send message to viable links
			for _, node := range links {
				m.outCh <- messages.NewMessage(messages.Packet, node.Address, data)
			}
		})
	*/
	return nil
}

func (m *Medium) getNodeByAddr(addr string) (*types.Node, error) {
	var node *types.Node
	for _, n := range *m.nodes {
		if n.Address == addr {
			node = &n
			break
		}
	}
	if node == nil {
		return nil, fmt.Errorf("no node found matching the provided address (%s)", addr)
	}

	return node, nil
}

func (m *Medium) getNodeAndBandInfo(address, band string) (*types.Node, *BandInfo, error) {
	nodeIndex := -1
	for i, n := range *m.nodes {
		if n.Address == address {
			nodeIndex = i
			break
		}
	}
	if nodeIndex < 0 {
		return nil, nil, fmt.Errorf("no node found matching the provided address (%s)", address)
	}

	bandIndex := -1
	for i, b := range m.bands[nodeIndex] {
		if b.Name == band {
			bandIndex = i
			break
		}
	}
	if bandIndex < 0 {
		return nil, nil, fmt.Errorf("no band found matching the provided address and band (%s:%s)", address, band)
	}

	return &(*m.nodes)[nodeIndex], &m.bands[nodeIndex][bandIndex], nil
}
