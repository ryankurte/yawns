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
	config config.Medium
	nodes  map[string]Node
	Links  []Link

	inCh  chan *messages.Message
	outCh chan *messages.Message
}

// NewMedium creates a new medium instance
func NewMedium(c *config.Config) *Medium {
	m := Medium{
		config: c.Medium,
		inCh:   make(chan *messages.Message, 128),
		outCh:  make(chan *messages.Message, 128),
	}

	m.nodes = make(map[string]Node)
	for _, node := range c.Nodes {
		m.nodes[node.Address] = Node{node, false, false}
	}

	m.Links = make([]Link, 0)

	// Iterate over nodes from config to ensure order is maintained
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

func (m *Medium) sendPacket(from string, data []byte) {

	// Set timeout for packet sent response
	packetTime := float64((len(data) + m.config.Overhead)) / m.config.Baud
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
		fading := l.Fading + GetRandomFading(m.config.Fading)

		// Drop links where fading is greater than the link budget
		if fading > m.config.LinkBudget {
			continue
		}

		// Add viable links to array
		if l.From == from {
			links = append(links, l.To)
		} else if l.To == from {
			links = append(links, l.From)
		}
	}

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

func (m *Medium) addNode(node *Node) {
	m.nodes[node.Address] = *node

	for addr, n := range m.nodes {
		if addr != node.Address {
			m.createLink(*node, n)
		}
	}
}

func (m *Medium) findLink(from, to string) *Link {
	for _, l := range m.Links {
		if l.From == from && l.To == to {
			return &l
		}
	}
	return nil
}

func (m *Medium) updateLink(link *Link) (*Link, error) {
	to, ok := m.nodes[link.To]
	if !ok {
		return nil, fmt.Errorf("Node not found (%s)", link.To)
	}
	from, ok := m.nodes[link.From]
	if !ok {
		return nil, fmt.Errorf("Node not found (%s)", link.From)
	}

	link.Distance = math.Floor(GetDistance(&from.Location, &to.Location))
	link.Fading = FreeSpaceAttenuationDB(Frequency(m.config.Frequency), Distance(link.Distance))

	return link, nil
}

func (m *Medium) createLink(a, b Node) {
	link := &Link{
		From: a.Address,
		To:   b.Address,
	}

	link, _ = m.updateLink(link)

	m.Links = append(m.Links, *link)

}
