package config

import (
	"github.com/ryankurte/owns/lib/types"
)

// Channels defines channel information for the medium
type Channels struct {
	// Number of channels
	Count uint64
	// Channel Spacing in Hz
	Spacing types.Frequency
}

// Band is a simulated frequency band
type Band struct {
	// Radio Frequency in Hz
	Frequency types.Frequency
	// Baud rate in bps
	Baud types.Baud
	// Packet overhead in bytes
	PacketOverhead types.SizeBytes
	// Standard deviation of gaussian fading in dB
	RandomDeviation types.Attenuation
	// Link Budget in dB
	LinkBudget types.Attenuation
	// Attenuation budget defines the minimum attenuation (in dB) at which signals will interfere (and cause packet corruption)
	InterferenceBudget types.Attenuation
	// Packet Error Rate
	ErrorRate float64
	// Channel information
	Channels Channels
	// Disable auto transition from tx to RX state
	NoAutoTXRXTransition bool
	// Noise floor in dB
	NoiseFloor types.Attenuation
}

// Maps configuration for the Medium Map layer
type Maps struct {
	// X and Y tile locations
	X, Y uint64
	// Map level
	Level uint64
	// Satellite map file
	Satellite string
	// Terrain map file
	Terrain string
	// Default terrain offset (for unset altitudes)
	DefaultOffset types.Distance
}

// Medium defines the simulator configuration for the medium module
type Medium struct {
	Maps      Maps
	Bands     map[string]Band // Frequency bands in simulation
	StatsFile string
}
