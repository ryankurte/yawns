package medium

import (
	"time"
)

type Stats struct {
	TickMin time.Duration
	TickMax time.Duration
	TickAvg time.Duration
	Bands   map[string]BandStats
	Nodes   map[string]NodeStats
	Links   []LinkStats
}

func NewStats() Stats {
	return Stats{
		TickMin: 0,
		TickMax: 0,
		TickAvg: 0,
		Bands:   make(map[string]BandStats),
		Nodes:   make(map[string]NodeStats),
		Links:   make([]LinkStats, 0),
	}
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

type LinkStats struct {
}
