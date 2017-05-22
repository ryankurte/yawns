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
	"container/list"

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

// Transmission is a transmission in flight
type Transmission struct {
	Origin     *types.Node
	Band       string
	Channel    int32
	Data       []byte
	StartTime  time.Time
	PacketTime time.Duration
	EndTime    time.Time
	SendOK     []bool
}

// Medium is the wireless medium simulation instance
type Medium struct {
	config        *config.Medium
	nodes         *[]types.Node
	transmissions *list.List

	layerManager *layers.LayerManager

	inCh  chan interface{}
	outCh chan interface{}
}

// NewMedium creates a new medium instance
func NewMedium(c *config.Medium, nodes *[]types.Node) *Medium {
	// Create base medium object
	m := Medium{
		config:        c,
		inCh:          make(chan interface{}, 128),
		outCh:         make(chan interface{}, 128),
		transmissions: list.New(),
		layerManager:  layers.NewLayerManager(),
		nodes:         nodes,
	}

	// Initialise RadioState for each node

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

func (m *Medium) sendPacket(p messages.Packet) error {

	fromAddress, bandName, channel, data := p.Address, p.Band, p.Channel, p.Data

	// Locate source node and source band info
	source, err := m.getNodeByAddr(fromAddress)
	if err != nil {
		return err
	}

	// Locate matching band
	band, ok := m.config.Bands[bandName]
	if !ok {
		return fmt.Errorf("Medium error: no matching band configured (%s)", bandName)
	}

	// Set transmitting state
	source.TransceiverStates[bandName] = types.TransceiverStateTransmitting

	// Calculate in flight time
	startTime := time.Now()
	packetTime := time.Duration(float64(len(data)+band.PacketOverhead) / float64(band.Baud) * float64(time.Second))
	endTime := startTime.Add(packetTime)

	// Create transmission instance
	t := Transmission{
		Origin:     source,
		Band:       bandName,
		Channel:    channel,
		Data:       data,
		StartTime:  startTime,
		PacketTime: packetTime,
		EndTime:    endTime,
		SendOK:     make([]bool, len(*m.nodes)),
	}

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

	fmt.Printf("Created transmission: %+v", t)

	m.transmissions.PushBack(t)

	return nil
}

func (m *Medium) update(now time.Time) {

	for e := m.transmissions.Front(); e != nil; e = e.Next() {
		t := e.Value.(Transmission)

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
			t.Origin.TransceiverStates[t.Band] = types.TransceiverStateIdle
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
			m.transmissions.Remove(e)
		}
	}

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
