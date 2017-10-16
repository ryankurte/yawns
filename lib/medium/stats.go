package medium

import (
	"math"
	"time"
)

// ContinuousFloat64 is a floating point number that continuously calculates statistics
// See: https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance
type ContinuousFloat64 struct {
	N      uint64
	Min    float64
	Max    float64
	Mean   float64
	StdDev float64

	mean2 float64
}

func NewContinuousFloat64() ContinuousFloat64 {
	return ContinuousFloat64{0, 0, 0, 0, 0, 0}
}

// Update adds a value to the continuous float and updates the computed statistics
func (c *ContinuousFloat64) Update(x float64) (n uint64, min, max, mean, stdDev float64) {
	// Update min and max
	if x > c.Max || c.N == 0 {
		c.Max = x
	}
	if x < c.Min || c.N == 0 {
		c.Min = x
	}

	// Update count
	c.N++

	// Update rolling standard deviation and mean
	delta := x - c.Mean
	c.Mean += delta / float64(c.N)
	delta2 := x - c.Mean
	c.mean2 += delta * delta2

	if c.N > 1 {
		c.StdDev = math.Sqrt(c.mean2 / float64(c.N-1))
	}

	return c.N, c.Min, c.Max, c.Mean, c.StdDev
}

// ContinuousDuration is a floating point number that continuously calculates statistics
// See: https://en.wikipedia.org/wiki/Algorithms_for_calculating_variance
type ContinuousDuration struct {
	cf     ContinuousFloat64
	N      uint64
	Min    time.Duration
	Max    time.Duration
	Mean   time.Duration
	StdDev time.Duration
}

// Update adds a value to the continuous float and updates the computed statistics
func (c *ContinuousDuration) Update(x time.Duration) {
	N, Min, Max, Mean, StdDev := c.cf.Update(float64(x))
	c.N, c.Min, c.Max, c.Mean, c.StdDev = N, time.Duration(Min), time.Duration(Max), time.Duration(Mean), time.Duration(StdDev)
}

type Stats struct {
	Tick  ContinuousDuration
	Bands map[string]BandStats
	Nodes map[string]NodeStats
	Links []LinkStats
}

func NewStats() Stats {
	return Stats{
		Tick:  ContinuousDuration{},
		Bands: make(map[string]BandStats),
		Nodes: make(map[string]NodeStats),
		Links: make([]LinkStats, 0),
	}
}

func (s *Stats) AddTick(t time.Duration) {
	s.Tick.Update(t)
}

type BandStats struct {
	PacketCount uint64
}

func (b *BandStats) IncrementPacketCount() {
	b.PacketCount++
}

type NodeStats struct {
	Sent     uint64
	Received uint64
}

func (n *NodeStats) IncrementReceived() {
	n.Received++
}

func (n *NodeStats) IncrementSent() {
	n.Sent++
}

type LinkStats struct {
}
