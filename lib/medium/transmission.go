/**
 * OpenNetworkSim Medium Package
 * Implements wireless medium simulation
 * Transmission implementation, this defines transmissions through the wireless medium
 *
 * https://github.com/ryankurte/ons
 * Copyright 2017 Ryan Kurte
 */

package medium

import (
	"github.com/ryankurte/owns/lib/config"
	"github.com/ryankurte/owns/lib/messages"
	"github.com/ryankurte/owns/lib/types"
	"time"
)

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
	RSSIs      [][]types.Attenuation
}

// NewTransmission creates a new transmission instance
func NewTransmission(now time.Time, origin *types.Node, band *config.Band, msg messages.Packet) *Transmission {

	// Calculate packet time for the given band
	packetTime := time.Duration(float64(len(msg.Data)+int(band.PacketOverhead)) * 8 / float64(band.Baud) * float64(time.Second))

	t := Transmission{
		Origin:     origin,
		Band:       msg.Band,
		Channel:    msg.Channel,
		Data:       msg.Data,
		StartTime:  now,
		PacketTime: packetTime,
		EndTime:    now.Add(packetTime),
	}
	return &t
}

func (t *Transmission) GetAverageRSSI(nodeIndex int) float64 {
	sum := types.Attenuation(0)
	for _, v := range t.RSSIs[nodeIndex] {
		sum += v
	}
	return float64(sum) / float64(len(t.RSSIs[nodeIndex]))
}

func (t *Transmission) GetRFInfo(nodeIndex int) messages.RFInfo {
	return messages.RFInfo{
		Band:    t.Band,
		Channel: t.Channel,
		RSSI:    t.GetAverageRSSI(nodeIndex),
	}
}

