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
	nodes         []*types.Node
	transmissions []*Transmission
	transceivers  []map[string]types.TransceiverState
	rate          time.Duration

	layerManager *layers.LayerManager

	inCh  chan interface{}
	outCh chan interface{}
}

// NewMedium creates a new medium instance
func NewMedium(c *config.Medium, rate time.Duration, nodes []*types.Node) *Medium {
	// Create base medium object
	m := Medium{
		config:        c,
		rate:          rate,
		inCh:          make(chan interface{}, 128),
		outCh:         make(chan interface{}, 128),
		transmissions: make([]*Transmission, 0),
		transceivers:  make([]map[string]types.TransceiverState, len(nodes)),
		layerManager:  layers.NewLayerManager(),
		nodes:         nodes,
	}

	// Initialise TransceiverState for each node and band
	for i := range nodes {
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
	for _, node := range m.nodes {
		// Skip source node
		if node.Address == source.Address {
			continue
		}

		// Calculate fading and add links where appropriate
		fading := m.GetPointToPointFading(band, source, *node)
		if band.LinkBudget > fading {
			visible = append(visible, node)
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
	source := m.nodes[nodeIndex]

	// Locate matching band
	band, ok := m.config.Bands[bandName]
	if !ok {
		return fmt.Errorf("Medium error: no matching band configured (%s)", bandName)
	}

	// Set transmitting state
	m.transceivers[nodeIndex][bandName] = types.TransceiverStateTransmitting

	// Create transmission instance
	t := NewTransmission(now, source, &band, p)
	t.SendOK = make([]bool, len(m.nodes))
	t.RSSIs = make([][]types.Attenuation, len(m.nodes))

	// Calculate initial transmission states for simulated nodes
	for i, n := range m.nodes {
		if n.Address == fromAddress {
			t.SendOK[i] = false
			continue
		}
		fading := m.GetPointToPointFading(band, *source, *n)
		t.RSSIs[i] = make([]types.Attenuation, 1)
		t.RSSIs[i][0] = fading
		if fading > band.LinkBudget {
			t.SendOK[i] = false
		} else {
			t.SendOK[i] = true
		}
	}

	// Add to transmission buffer
	m.transmissions = append(m.transmissions, t)

	return nil
}

func (m *Medium) update(now time.Time) {
	// Update in flight transmissions
	for i, t := range m.transmissions {
		// Update receive states
		band := m.config.Bands[t.Band]
		for j, n := range m.nodes {
			if n.Address == t.Origin.Address {
				continue
			}
			fading := m.GetPointToPointFading(band, *t.Origin, *n)
			m.transmissions[i].RSSIs[j] = append(t.RSSIs[j], fading)
			if t.SendOK[j] && fading > band.LinkBudget {
				log.Printf("Updating failed state for node %d (%s)", j, n.Address)
				m.transmissions[i].SendOK[j] = false
			}
		}
	}

	// Calculate collisions for each node
	for i, n := range m.nodes {
		// Compare all transmissions
		for j1, t1 := range m.transmissions {
			for j2, t2 := range m.transmissions {
				if j1 == j2 || t1.Band != t2.Band || t1.Channel != t2.Channel ||
					n.Address == t1.Origin.Address || n.Address == t2.Origin.Address {
					continue
				}
				// Don't worry if both already failed
				if !t1.SendOK[i] && !t2.SendOK[i] {
					continue
				}

				// RSSI difference calculated on last saved RSSI from previous update stage
				rssiDifference := t1.RSSIs[i][len(t1.RSSIs[i])-1] - t2.RSSIs[i][len(t2.RSSIs[i])-1]
				band := m.config.Bands[t1.Band]

				// If difference is less than the interference budget, fail at sending both
				if (rssiDifference > 0 && rssiDifference < band.InterferenceBudget) ||
					(rssiDifference < 0 && rssiDifference > -band.InterferenceBudget) {
					m.transmissions[j1].SendOK[i] = false
					m.transmissions[j2].SendOK[i] = false
				}
			}
		}
	}

	// Complete sending after timeout
	toRemove := make([]int, 0)
	for i, t := range m.transmissions {
		if now.After(t.EndTime) {
			sourceIndex, _ := m.getNodeIndex(t.Origin.Address)

			// Update origin transmitting state
			m.transceivers[sourceIndex][t.Band] = types.TransceiverStateTransmitting
			m.outCh <- messages.NewSendComplete(t.Origin.Address, t.Band, t.Channel)

			// Distribute to receivers
			for i, n := range m.nodes {
				if t.SendOK[i] {
					m.outCh <- messages.NewPacket(n.Address, t.Data, t.GetRFInfo(i))
				}
			}

			// Remove from transmission list
			toRemove = append(toRemove, i)
		}
	}

	// Remove completed transmissions
	removedCount := 0
	for _, v := range toRemove {
		index := v - removedCount
		m.transmissions = append(m.transmissions[:index], m.transmissions[index+1:]...)
		removedCount++
	}
}

func (m *Medium) getNodeIndex(addr string) (int, error) {
	for i, n := range m.nodes {
		if n.Address == addr {
			return i, nil
		}
	}
	return 0, fmt.Errorf("no node found matching the provided address (%s)", addr)
}

func (m *Medium) getNodeByAddr(addr string) (*types.Node, error) {
	var node *types.Node
	for _, n := range m.nodes {
		if n.Address == addr {
			node = n
			break
		}
	}
	if node == nil {
		return nil, fmt.Errorf("no node found matching the provided address (%s)", addr)
	}

	return node, nil
}
