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
	"github.com/ryankurte/ons/lib/config"
	"github.com/ryankurte/ons/lib/messages"
	"github.com/ryankurte/ons/lib/types"
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
}

// NewTransmission creates a new transmission instance
func NewTransmission(now time.Time, origin *types.Node, band *config.Band, msg messages.Packet) Transmission {

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
	return t
}
