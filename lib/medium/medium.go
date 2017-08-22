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

	"github.com/ryankurte/owns/lib/config"
	"github.com/ryankurte/owns/lib/medium/layers"
	"github.com/ryankurte/owns/lib/messages"
	"github.com/ryankurte/owns/lib/types"
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
	transmissions []*Transmission
	transceivers  []map[string]Transceiver
	rate          time.Duration

	layerManager *layers.LayerManager

	stats Stats

	inCh  chan interface{}
	outCh chan interface{}
}

type Stats struct {
	TickMin time.Duration
	TickMax time.Duration
	TickAvg time.Duration
}

// NewMedium creates a new medium instance
func NewMedium(c *config.Medium, rate time.Duration, nodes *[]types.Node) (*Medium, error) {
	// Create base medium object
	m := Medium{
		config:        c,
		rate:          rate,
		inCh:          make(chan interface{}, 128),
		outCh:         make(chan interface{}, 128),
		transmissions: make([]*Transmission, 0),
		transceivers:  make([]map[string]Transceiver, len(*nodes)),
		layerManager:  layers.NewLayerManager(),
		nodes:         nodes,
	}

	// Initialise TransceiverState for each node and band
	for i := range *nodes {
		m.transceivers[i] = make(map[string]Transceiver)
		for j := range c.Bands {
			m.transceivers[i][j] = *NewTransceiver(time.Now())
		}
	}

	// Load medium simulation layers
	m.layerManager.BindLayer(layers.NewFreeSpace())
	m.layerManager.BindLayer(layers.NewRandom())

	mapLayer, err := layers.NewMapLayer(&c.Maps)
	if err != nil {
		return nil, err
	}
	m.layerManager.BindLayer(mapLayer)

	return &m, nil
}

func (m *Medium) Send() chan interface{} {
	return m.inCh
}

func (m *Medium) Receive() chan interface{} {
	return m.outCh
}

// GetPointToPointFading fetches the (instantaneous) fading between two nodes at a given frequency
func (m *Medium) GetPointToPointFading(band config.Band, n1, n2 types.Node) types.Attenuation {
	return types.Attenuation(m.layerManager.CalculateFading(band, n1.Location, n2.Location))
}

func (m *Medium) Start() {
	go m.Run()
}

func (m *Medium) Stop() {
	close(m.inCh)
	close(m.outCh)
}

// Run runs the medium simulation
func (m *Medium) Run() {
	log.Printf("[INFO] Medium running")

	m.stats.TickMin = 0
	m.stats.TickMax = 0
	m.stats.TickAvg = 0

	var avg float64
	var count float64

	lastTime := time.Now()
	runTimer := time.NewTicker(m.rate)

running:
	for {
		select {
		case message, ok := <-m.inCh:
			if !ok {
				log.Printf("[INFO] Medium input channel closed")
				break running
			}

			err := m.handleMessage(message)
			if err != nil {
				log.Printf("[ERROR] Medium error: %s", err)
			}

		// Run timed updates
		case now := <-runTimer.C:
			// Calculate delta between runs
			if count != 0 {
				delta := now.Sub(lastTime)
				lastTime = now
				if m.stats.TickMax == 0 || delta > m.stats.TickMax {
					m.stats.TickMax = delta
				}
				if m.stats.TickMin == 0 || delta < m.stats.TickMin {
					m.stats.TickMin = delta
				}
				avg = avg*count/(count+1) + float64(delta)/(count+1)
			}
			count++
			m.update(now)
		}
	}

	m.stats.TickAvg = time.Duration(avg)
	log.Printf("[INFO] Medium exited (stats: %+v)", m.stats)
}

func (m *Medium) handleMessage(message interface{}) error {
	switch msg := message.(type) {
	case *messages.Packet:
		return m.sendPacket(time.Now(), *msg)
	case *messages.RSSIRequest:
		rssi, _ := m.getRSSI(msg.Address, msg.Band, msg.Channel)
		m.outCh <- &messages.RSSIResponse{
			BaseMessage: msg.BaseMessage,
			RFInfo:      msg.RFInfo,
			RSSI:        float32(rssi),
		}
	case *messages.StartReceive:
		m.setTransceiverState(msg.Address, msg.Band, TransceiverStateReceive)

	case *messages.StopReceive:
		m.setTransceiverState(msg.Address, msg.Band, TransceiverStateIdle)

	default:
		log.Printf("[WARNING] medium unhandled message type: %+t", message)
	}

	return nil
}

func (m *Medium) setTransceiverState(address, band string, state TransceiverState) error {
	index, err := m.getNodeIndex(address)
	if err != nil {
		return err
	}
	transceiver, ok := m.transceivers[index][band]
	if !ok {
		return fmt.Errorf("Transceiver not found for band: %s", band)
	}
	transceiver.SetState(time.Now(), state)
	m.transceivers[index][band] = transceiver
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

	log.Printf("[DEBUG] Medium - Starting transmission from %s", p.Address)

	// Set transmitting state
	m.SetTransceiverState(now, nodeIndex, bandName, TransceiverStateTransmitting)

	// Create transmission instance
	t := NewTransmission(now, source, &band, p)
	t.SendOK = make([]bool, len(*m.nodes))
	t.RSSIs = make([][]types.Attenuation, len(*m.nodes))

	// Calculate initial transmission states for simulated nodes
	for i, n := range *m.nodes {
		if n.Address == fromAddress {
			t.SendOK[i] = false
			continue
		}

		fading := m.GetPointToPointFading(band, *source, n)
		t.RSSIs[i] = make([]types.Attenuation, 1)
		t.RSSIs[i][0] = fading

		// Reject if fading exceeds link budget
		if fading > band.LinkBudget {
			t.SendOK[i] = false
			continue
		}

		t.SendOK[i] = true

		// Update radio states
		transceiver := m.transceivers[i][t.Band]
		if transceiver.State == TransceiverStateReceive {
			// Devices in receive state will enter receiving state
			m.SetTransceiverState(now, i, t.Band, TransceiverStateReceiving)
		}
	}

	// Add to transmission buffer
	m.transmissions = append(m.transmissions, t)

	return nil
}

// update updates the wireless medium simulation
func (m *Medium) update(now time.Time) {
	// Update in flight transmissions
	m.updateTransmissions(now)

	// Calculate collisions for each pair of transmissions at each node
	m.updateCollisions(now)

	// Finalise completed transmissions
	m.finaliseTransmissions(now)
}

// updateTransmissions updates a transmission RSSI and fading limits
func (m *Medium) updateTransmissions(now time.Time) {
	// Update in flight transmissions
	for i, t := range m.transmissions {
		// Update receive states
		band := m.config.Bands[t.Band]
		for j, n := range *m.nodes {
			if n.Address == t.Origin.Address {
				continue
			}
			fading := m.GetPointToPointFading(band, *t.Origin, n)
			m.transmissions[i].RSSIs[j] = append(t.RSSIs[j], fading)

			// Reject if fading exceeds link budget
			if t.SendOK[j] && fading > band.LinkBudget {
				log.Printf("Updating failed state for node %d (%s)", j, n.Address)
				m.transmissions[i].SendOK[j] = false
				m.SetTransceiverState(now, i, t.Band, TransceiverStateReceive)
			}

			// TODO: Reject if radio exits receiving state
			// state := m.transceivers[j][t.Band]
			/*
				if t.SendOK[j] && state != TransceiverStateReceiving {
					// If the device is not in the receiving state fail
					m.transmissions[i].SendOK[j] = false
				}
			*/
		}
	}
}

// updateCollisions calculates collisions based on the interference budget and last rssi value
func (m *Medium) updateCollisions(now time.Time) {
	for i, n := range *m.nodes {
		// Compare all transmissions
		for j1, t1 := range m.transmissions {
			for j2, t2 := range m.transmissions {
				// Filter transmissions we don't need to compare
				if j1 == j2 || t1.Band != t2.Band || t1.Channel != t2.Channel ||
					n.Address == t1.Origin.Address || n.Address == t2.Origin.Address ||
					(!t1.SendOK[i] && !t2.SendOK[i]) {
					continue
				}

				// RSSI difference calculated on last saved RSSI from previous update stage
				rssiDifference := t1.RSSIs[i][len(t1.RSSIs[i])-1] - t2.RSSIs[i][len(t2.RSSIs[i])-1]
				band := m.config.Bands[t1.Band]

				// If difference is less than the interference budget, fail at sending both
				if (rssiDifference > 0 && rssiDifference < band.InterferenceBudget) ||
					(rssiDifference < 0 && rssiDifference > -band.InterferenceBudget) {
					log.Printf("Updating collision state for node %d (%s)", i, n.Address)
					m.transmissions[j1].SendOK[i] = false
					m.transmissions[j2].SendOK[i] = false
					m.SetTransceiverState(now, i, t1.Band, TransceiverStateReceive)
				}
			}
		}
	}
}

func (m *Medium) getRSSI(address, bandName string, channel int32) (types.Attenuation, error) {
	band := m.config.Bands[bandName]
	nodeIndex, err := m.getNodeIndex(address)
	if err != nil {
		return 0.0, err
	}

	rssi := band.NoiseFloor
	for _, t := range m.transmissions {
		if t.Band != bandName || t.Channel != channel {
			continue
		}
		fading := t.RSSIs[nodeIndex][len(t.RSSIs[nodeIndex])-1]

		if fading > rssi {
			rssi = fading
		}

	}
	return rssi, nil
}

// finaliseTransmissions finalises any completed transmissions
func (m *Medium) finaliseTransmissions(now time.Time) {
	// Complete sending after timeout
	toRemove := make([]int, 0)
	for i, t := range m.transmissions {
		if now.After(t.EndTime) {
			sourceIndex, _ := m.getNodeIndex(t.Origin.Address)
			band := m.config.Bands[t.Band]

			log.Printf("[DEBUG] Medium - Completing transmission from %s", t.Origin.Address)

			// Update origin transmitting state
			m.outCh <- messages.NewSendComplete(t.Origin.Address, t.Band, t.Channel)
			if band.NoAutoTXRXTransition {
				m.SetTransceiverState(now, sourceIndex, t.Band, TransceiverStateIdle)
			} else {
				m.SetTransceiverState(now, sourceIndex, t.Band, TransceiverStateReceive)
			}

			// Distribute to receivers
			for i, n := range *m.nodes {
				if t.SendOK[i] {
					m.outCh <- messages.NewPacket(n.Address, t.Data, t.GetRFInfo(i))
					m.SetTransceiverState(now, i, t.Band, TransceiverStateReceive)
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
	for i, n := range *m.nodes {
		if n.Address == addr {
			return i, nil
		}
	}
	return 0, fmt.Errorf("no node found matching the provided address (%s)", addr)
}

func (m *Medium) getNodeByAddr(addr string) (*types.Node, error) {
	var node *types.Node
	for i, n := range *m.nodes {
		if n.Address == addr {
			node = &(*m.nodes)[i]
			break
		}
	}
	if node == nil {
		return nil, fmt.Errorf("no node found matching the provided address (%s)", addr)
	}

	return node, nil
}
