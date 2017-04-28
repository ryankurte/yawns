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
	"io/ioutil"
	"log"
	"math"
	"time"

	"github.com/ryankurte/ons/lib/config"
	"github.com/ryankurte/ons/lib/medium/layers"
	"github.com/ryankurte/ons/lib/messages"
	"gopkg.in/yaml.v2"
)

// Link encapsulates a link between two nodes
type Link struct {
	From, To string
	Distance float64
	Fading   float64
}

type Node struct {
	config.Node
	Transmitting bool
	Receiving    bool
}

// Medium is the wireless medium simulation instance
type Medium struct {
	config *config.Medium
	nodes  *[]Node
	Links  []Link

	layerManager *layers.LayerManager

	inCh  chan *messages.Message
	outCh chan *messages.Message
}

// NewMedium creates a new medium instance
func NewMedium(c *config.Medium) *Medium {
	// Create base medium object
	m := Medium{
		config:       c,
		inCh:         make(chan *messages.Message, 128),
		outCh:        make(chan *messages.Message, 128),
		layerManager: layers.NewLayerManager(),
	}

	// Load medium simulation layers
	m.layerManager.BindLayer(layers.NewFreeSpace())
	m.layerManager.BindLayer(layers.NewRandom(float64(c.Medium.RandomDeviation)))

	// Load nodes from configuration
	// Note this is iterated to maintain config file order
	for _, a := range c.Nodes {
		for _, b := range c.Nodes {
			if a.Address != b.Address {
				if link := m.findLink(b.Address, a.Address); link != nil {
					continue
				}

				log.Printf("Creating link from %s to %s", a.Address, b.Address)
				m.createLink(m.nodes[a.Address], m.nodes[b.Address])
			}
		}
	}

	log.Printf("Links: %+v", m.Links)

	return &m
}

func (m *Medium) WriteYML(file string) error {
	data, err := yaml.Marshal(m)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(file, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (m *Medium) GetVisible(from string) []string {
	visible := make([]string, 0)

	for _, link := range m.Links {
		if link.From == from {
			visible = append(visible, link.To)
		}
		if link.To == from {
			visible = append(visible, link.From)
		}
	}

	return visible
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

func (m *Medium) handleMessage(message *messages.Message) {
	switch message.GetType() {
	case messages.Packet:
		m.sendPacket(message.GetAddress(), message.GetData())
	default:
		log.Printf("Medium unhandled message: %+v", message)
	}

}

// CalculateFading calculates the fading between two points using the available layers
func (m *Medium) CalculateFading(freq float64, p1, p2 config.Location) float64 {
	fading := 0.0
	for _, layer := range m.layers {
		fading += layer.CalculateFading(freq, p1, p2)
	}
	return fading
}

func (m *Medium) sendPacket(from string, data []byte) {

	// Set timeout for packet sent response
	packetTime := float64((len(data) + m.config.Overhead)) / float64(m.config.Baud)
	if node, ok := m.nodes[from]; ok {
		node.Transmitting = true
	}

	// Build a list of viable links
	links := make([]string, 0)
	for _, l := range m.Links {

		// Skip irrelevant links
		if l.From != from && l.To != from {
			continue
		}

		// Calculate fading (Free space + random)
		// TODO: this should one day include fresnel zone impingement
		// Perhaps this could be implemented as fading layers..?
		fading := l.Fading + GetRandomFading(float64(m.config.Fading))

		// Drop links where fading is greater than the link budget
		if fading > float64(m.config.LinkBudget) {
			continue
		}

		// Add viable links to array
		if l.From == from {
			links = append(links, l.To)
		} else if l.To == from {
			links = append(links, l.From)
		}
	}

	fmt.Printf("Viable links: %+v", links)

	// Run callback after packet send has completed
	time.AfterFunc(time.Duration(packetTime)*time.Second, func() {
		// Send packet-sent message to application
		if node, ok := m.nodes[from]; ok {
			node.Transmitting = false
			m.outCh <- messages.NewMessage(messages.PacketSent, from, []byte{})
		}

		// Send message to viable links
		for _, l := range links {
			m.outCh <- messages.NewMessage(messages.Packet, l, data)
		}
	})
}
