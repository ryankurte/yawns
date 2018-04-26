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
	"time"

	"github.com/ryankurte/owns/lib/config"
	"github.com/ryankurte/owns/lib/helpers"
	"github.com/ryankurte/owns/lib/medium/layers"
	"github.com/ryankurte/owns/lib/messages"
	"github.com/ryankurte/owns/lib/types"
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
	nodes         *types.Nodes
	transmissions []*Transmission
	transceivers  []map[string]Transceiver
	rate          time.Duration

	layerManager *layers.LayerManager

	stats Stats

	inCh  chan interface{}
	outCh chan interface{}
}

// NewMedium creates a new medium instance
func NewMedium(c *config.Medium, rate time.Duration, nodes *types.Nodes) (*Medium, error) {
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
		stats:         NewStats(),
	}

	// Initialise TransceiverState for each node and band
	for i, n := range *nodes {
		m.stats.Nodes[n.Address] = NewNodeStats()
		m.transceivers[i] = make(map[string]Transceiver)
		for j := range c.Bands {
			m.transceivers[i][j] = *NewTransceiver(time.Now())
		}
	}

	m.BindDefaultLayers(c)

	return &m, nil
}

// BindDefaultLayers Binds default medium layers
func (m *Medium) BindDefaultLayers(c *config.Medium) error {
	// Load medium simulation layers
	m.layerManager.BindLayer("free-space", layers.NewFreeSpace())
	m.layerManager.BindLayer("random", layers.NewRandom())

	if c.Maps.Satellite != "" {
		mapLayer, err := layers.NewRenderLayer(&c.Maps)
		if err != nil {
			return err
		}
		m.layerManager.BindLayer("render", mapLayer)
	}

	if c.Maps.Terrain != "" {
		mapLayer, err := layers.NewTerrainLayer(&c.Maps)
		if err != nil {
			return err
		}
		m.layerManager.BindLayer("terrain", mapLayer)
	}

	if c.Maps.Foliage != "" {
		mapLayer, err := layers.NewFoliageLayer(&c.Maps)
		if err != nil {
			return err
		}
		m.layerManager.BindLayer("foliage", mapLayer)
	}

	return nil
}

func (m *Medium) BindLayer(name string, layer interface{}) error {
	return m.layerManager.BindLayer(name, layer)
}

func (m *Medium) GetLayer(name string) (interface{}, error) {
	return m.layerManager.GetLayer(name)
}

func (m *Medium) Send() chan interface{} {
	return m.inCh
}

func (m *Medium) Receive() chan interface{} {
	return m.outCh
}

func (m *Medium) preloadFadings() {
	for _, v := range m.config.Bands {
		for i1, n1 := range *m.nodes {
			for i2, n2 := range *m.nodes {
				if i1 == i2 {
					continue
				}
				m.GetPointToPointFading(v, n1, n2)
			}
		}
	}
}

// GetPointToPointFading fetches the (instantaneous) fading between two nodes at a given frequency
func (m *Medium) GetPointToPointFading(band config.Band, n1, n2 types.Node) types.AttenuationMap {
	attenuation, err := m.layerManager.CalculateFading(band, n1.Location, n2.Location)
	if err != nil {
		log.Printf("[ERROR] medium layer error: %s", err)
	}

	return attenuation
}

func (m *Medium) Start() {
	m.preloadFadings()

	go m.Run()
}

func (m *Medium) Stop() {
	close(m.inCh)
}

// Run runs the medium simulation
func (m *Medium) Run() {
	log.Printf("[INFO] Medium running")

	lastTime := time.Now()
	runTimer := time.NewTicker(m.rate)

running:
	for {
		select {
		case message, ok := <-m.inCh:
			if !ok {
				log.Printf("[INFO] Medium input channel closed")
				close(m.outCh)
				break running
			}

			err := m.handleMessage(message)
			if err != nil {
				log.Printf("[ERROR] Medium error: %s", err)
			}

		// Run timed updates
		case now := <-runTimer.C:
			// Calculate delta between runs
			delta := now.Sub(lastTime)
			lastTime = now
			m.stats.AddTick(delta)
			m.update(now)
		}
	}

	for i, n := range *m.nodes {
		for k, t := range m.transceivers[i] {
			m.stats.Nodes[n.Address].Transceivers[k] = t.Stats
		}
	}

	log.Printf("[INFO] Medium exited")
	log.Printf("Medium Info")
	log.Printf("  - Tick Count: %d Mean: %s StdDev: %s Min: %s Max: %s", m.stats.Tick.Count, m.stats.Tick.Mean, m.stats.Tick.StdDev, m.stats.Tick.Min, m.stats.Tick.Max)

	if m.config.StatsFile != "" {
		helpers.WriteYAMLFile(m.config.StatsFile, &m.stats)
	}

	runTimer.Stop()
}

func (m *Medium) handleMessage(message interface{}) error {
	switch msg := message.(type) {
	case messages.Packet:
		return m.sendPacket(time.Now(), msg)
	case messages.RSSIRequest:
		rssi, err := m.getRSSI(msg.Address, msg.Band, msg.Channel)
		if err != nil {
			log.Printf("[ERROR] Medium RSSI get error: %s", err)
		}
		m.outCh <- messages.RSSIResponse{
			BaseMessage: msg.BaseMessage,
			RFInfo:      msg.RFInfo,
			RSSI:        float32(rssi),
		}
	case messages.StateRequest:
		nodeIndex, _ := m.getNodeIndex(msg.Address)
		state := m.transceivers[nodeIndex][msg.Band].State
		m.outCh <- messages.StateResponse{
			BaseMessage: msg.BaseMessage,
			RFInfo:      msg.RFInfo,
			State:       state,
		}

	case messages.StateSet:
		m.setTransceiverState(msg.Address, msg.Band, msg.State)

	case messages.Register:
		// Mock to avoid warning on unhandled message

	case messages.FieldSet:
		// Mock to avoid warning on unhandled message

	default:
		log.Printf("[WARNING] medium unhandled message type: %T", message)
	}

	return nil
}

func (m *Medium) Render(filename string, nodes types.Nodes, links types.Links) error {
	return m.layerManager.Render(filename, nodes, links)
}

func (m *Medium) setTransceiverState(address, band string, state types.TransceiverState) error {
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

	m.stats.IncrementSent(fromAddress, bandName)

	//log.Printf("[DEBUG] Medium - Starting transmission from %s", fromAddress)

	// Set transmitting state
	m.setTransceiverState(fromAddress, bandName, types.TransceiverStateTransmitting)

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

		fading := m.GetPointToPointFading(band, *source, n).Reduce()
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
		if transceiver.State == types.TransceiverStateReceive {
			// Devices in receive state will enter receiving state
			m.setTransceiverState(n.Address, bandName, types.TransceiverStateReceiving)
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
			fading := m.GetPointToPointFading(band, *t.Origin, n).Reduce()
			m.transmissions[i].RSSIs[j] = append(t.RSSIs[j], fading)

			// Reject if fading exceeds link budget
			if t.SendOK[j] && fading > band.LinkBudget {
				log.Printf("Updating failed state for node %d (%s)", j, n.Address)
				m.transmissions[i].SendOK[j] = false
				m.setTransceiverState(n.Address, t.Band, types.TransceiverStateReceive)
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

				if !m.transmissions[j1].SendOK[i] || !m.transmissions[j2].SendOK[i] {
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
					m.setTransceiverState(n.Address, t1.Band, types.TransceiverStateReceive)
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
			band := m.config.Bands[t.Band]

			//log.Printf("[DEBUG] Medium - Completing transmission from %s", t.Origin.Address)

			// Update origin transmitting state
			m.outCh <- messages.NewSendComplete(t.Origin.Address, t.Band, t.Channel)
			if band.NoAutoTXRXTransition {
				m.setTransceiverState(t.Origin.Address, t.Band, types.TransceiverStateIdle)
			} else {
				m.setTransceiverState(t.Origin.Address, t.Band, types.TransceiverStateReceive)
			}

			// Distribute to receivers
			for i, n := range *m.nodes {
				if t.SendOK[i] && m.transceivers[i][t.Band].State == types.TransceiverStateReceiving {
					m.outCh <- messages.NewPacket(n.Address, t.Data, t.GetRFInfo(i))
					m.setTransceiverState(n.Address, t.Band, types.TransceiverStateReceive)
					m.stats.IncrementReceived(t.Origin.Address, n.Address, t.Band)
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
