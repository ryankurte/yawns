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
	"time"
)

// Link encapsulates a link between two nodes
type Link struct {
	From, To string
	Distance float64
	Fading   float64
}

// Medium is the wireless medium simulation instance
type Medium struct {
	config        *config.Medium
	nodes         *[]types.Node
	transmissions []Transmission
	transceivers  []map[string]types.TransceiverState
	rate          time.Duration

	layerManager *layers.LayerManager

	inCh  chan interface{}
	outCh chan interface{}
}

// NewMedium creates a new medium instance
func NewMedium(c *config.Medium, rate time.Duration, nodes *[]types.Node) *Medium {
	// Create base medium object
	m := Medium{
		config:        c,
		rate:          rate,
		inCh:          make(chan interface{}, 128),
		outCh:         make(chan interface{}, 128),
		transmissions: make([]Transmission, 0),
		transceivers:  make([]map[string]types.TransceiverState, len(*nodes)),
		layerManager:  layers.NewLayerManager(),
		nodes:         nodes,
	}

	// Initialise TransceiverState for each node and band
	for i := range *nodes {
		m.transceivers[i] = make(map[string]types.TransceiverState)
		for j := range c.Bands {
			m.transceivers[i][j] = types.TransceiverStateIdle
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

// Run runs the medium simulation
func (m *Medium) Run() {
	log.Printf("Medium running")
	runTimer := time.NewTicker(m.rate)
running:
	for {
		select {
		case message, ok := <-m.inCh:
			if !ok {
				log.Printf("Medium input channel closed")
				break running
			}

			err := m.handleMessage(message)
			if err != nil {
				log.Printf("Medium error: %s", err)
			}
		case now := <-runTimer.C:
			m.update(now)
		}
	}
	log.Printf("Medium exited")
}

func (m *Medium) handleMessage(message interface{}) error {
	switch msg := message.(type) {
	case messages.Packet:
		log.Printf("Received packet: %+v", msg)
		return m.sendPacket(time.Now(), msg)
	default:
		log.Printf("Medium unhandled message: %+v", message)
	}
	return nil
}

func (m *Medium) sendPacket(now time.Time, p messages.Packet) error {

	fromAddress, bandName := p.Address, p.Band

	// Locate source node and source band info
	nodeIndex, err := m.getNodeIndex(p.Address)
	if err != nil {
		return err
	}
	source := &(*m.nodes)[nodeIndex]

	// Locate matching band
	band, ok := m.config.Bands[bandName]
	if !ok {
		return fmt.Errorf("Medium error: no matching band configured (%s)", bandName)
	}

	// Set transmitting state
	m.transceivers[nodeIndex][bandName] = types.TransceiverStateTransmitting

	// Create transmission instance
	t := NewTransmission(now, source, &band, p)
	t.SendOK = make([]bool, len(*m.nodes))

	// Calculate initial transmission states for simulated nodes
	for i, n := range *m.nodes {
		if n.Address == fromAddress {
			t.SendOK[i] = false
			continue
		}
		fading := m.GetPointToPointFading(band, *source, n)
		if fading > band.LinkBudget {
			t.SendOK[i] = false
		} else {
			t.SendOK[i] = true
		}
	}

	fmt.Printf("Created transmission: %+v\n", t)

	m.transmissions = append(m.transmissions, t)

	fmt.Printf("Medium 1: %+v\n", &m)

	return nil
}

func (m *Medium) update(now time.Time) {

	for i, t := range m.transmissions {
		sourceIndex, _ := m.getNodeIndex(t.Origin.Address)

		// Locate matching band
		band := m.config.Bands[t.Band]

		// Update receive states
		for i, n := range *m.nodes {
			if n.Address == t.Origin.Address {
				continue
			}
			fading := m.GetPointToPointFading(band, *t.Origin, n)
			if fading > band.LinkBudget {
				t.SendOK[i] = false
			}
		}

		// Complete sending after timeout
		if t.EndTime.After(now) {
			// Update origin transmitting state
			m.transceivers[sourceIndex][t.Band] = types.TransceiverStateTransmitting
			m.outCh <- &messages.SendComplete{
				Message: messages.Message{
					Address: t.Origin.Address,
				},
				RFInfo: messages.NewRFInfo(t.Band, t.Channel),
			}

			// Distribute to receivers
			for i, n := range *m.nodes {
				if t.SendOK[i] {
					m.outCh <- &messages.Packet{
						Message: messages.Message{
							Address: n.Address,
						},
						Data: t.Data,
						RFInfo: messages.RFInfo{
							Band:    t.Band,
							Channel: t.Channel,
						},
					}
				}
			}

			// remove from transmission list
			m.transmissions = append(m.transmissions[:i], m.transmissions[i+1:]...)
		}
	}

}

func (m *Medium) getNodeIndex(addr string) (int, error) {
	for i, n := range *m.nodes {
		if n.Address == addr {
			return i, nil
		}
	}
	return 0, fmt.Errorf("no node found matching the provided address (%s)", addr)
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
