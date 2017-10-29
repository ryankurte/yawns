package medium

import (
	"math"
	"time"
)

// continuousFloat64 is a floating point number that continuously calculates statistics
// See: https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance
type continuousFloat64 struct {
	Count  uint64
	Min    float64
	Max    float64
	Mean   float64
	StdDev float64

	mean2 float64
}

func NewcontinuousFloat64() continuousFloat64 {
	return continuousFloat64{0, 0, 0, 0, 0, 0}
}

// Update adds a value to the continuous float and updates the computed statistics
func (c *continuousFloat64) Update(x float64) (n uint64, min, max, mean, stdDev float64) {
	// Update min and max
	if x > c.Max || c.Count == 0 {
		c.Max = x
	}
	if x < c.Min || c.Count == 0 {
		c.Min = x
	}

	// Update count
	c.Count++

	// Update rolling standard deviation and mean
	delta := x - c.Mean
	c.Mean += delta / float64(c.Count)
	delta2 := x - c.Mean
	c.mean2 += delta * delta2

	if c.Count > 1 {
		c.StdDev = math.Sqrt(c.mean2 / float64(c.Count-1))
	}

	return c.Count, c.Min, c.Max, c.Mean, c.StdDev
}

// continuousDuration is a floating point number that continuously calculates statistics
// See: https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance
type continuousDuration struct {
	cf     continuousFloat64
	Count  uint64
	Min    time.Duration
	Max    time.Duration
	Mean   time.Duration
	StdDev time.Duration
}

// Update adds a value to the continuous float and updates the computed statistics
func (c *continuousDuration) Update(x time.Duration) {
	N, Min, Max, Mean, StdDev := c.cf.Update(float64(x))
	c.Count, c.Min, c.Max, c.Mean, c.StdDev = N, time.Duration(Min), time.Duration(Max), time.Duration(Mean), time.Duration(StdDev)
}

type Stats struct {
	Tick  continuousDuration
	Bands map[string]BandStats
	Nodes map[string]NodeStats
	Links map[string][]LinkStats
}

func NewStats() Stats {
	return Stats{
		Tick:  continuousDuration{},
		Bands: make(map[string]BandStats),
		Nodes: make(map[string]NodeStats),
		Links: make(map[string][]LinkStats, 0),
	}
}

func (s *Stats) AddTick(t time.Duration) {
	s.Tick.Update(t)
}

func (s *Stats) IncrementSent(address string, band string) {
	nodeStats, ok := s.Nodes[address]
	if !ok {
		nodeStats = NewNodeStats()
	}
	nodeStats.IncrementSent()
	s.Nodes[address] = nodeStats

	bandStats, ok := s.Bands[band]
	if !ok {
		bandStats = NewBandStats()
	}
	bandStats.PacketCount++
	s.Bands[band] = bandStats
}

func (s *Stats) IncrementReceived(from, to string, band string) {
	nodeStats, ok := s.Nodes[to]
	if !ok {
		nodeStats = NewNodeStats()
	}

	nodeStats.IncrementReceived()
	s.Nodes[to] = nodeStats

	found := false
	for i, l := range s.Links[band] {
		if l.From == from && l.To == to {
			l.Sent++
			s.Links[band][i] = l
			found = true
		}
	}
	if !found {
		s.Links[band] = append(s.Links[band], LinkStats{From: from, To: to, Sent: 1})
	}
}

type BandStats struct {
	PacketCount uint64
}

func NewBandStats() BandStats {
	return BandStats{}
}

func (b *BandStats) IncrementPacketCount() {
	b.PacketCount++
}

type TransceiverStats struct {
	// Time spent with transcever off
	OffTime time.Duration
	// Time spent in idle mode
	IdleTime time.Duration
	// Time spent in sleep mode
	SleepTime time.Duration
	// Time spent in receive (listening) mode
	ReceiveTime time.Duration
	// Time spent receiving packets
	ReceivingTime time.Duration
	// Time spent transmitting packets
	TransmittingTime time.Duration
}

type NodeStats struct {
	Sent         uint64
	Received     uint64
	Transceivers map[string]TransceiverStats
}

func NewNodeStats() NodeStats {
	return NodeStats{
		Sent:         0,
		Received:     0,
		Transceivers: make(map[string]TransceiverStats),
	}
}

func (n *NodeStats) IncrementReceived() {
	n.Received++
}

func (n *NodeStats) IncrementSent() {
	n.Sent++
}

type LinkStats struct {
	To   string
	From string
	Sent uint64
}

func NewLinkStats() LinkStats {
	return LinkStats{}
}
