package config

import (
	"github.com/ryankurte/ons/lib/types"
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
}

// Medium defines the simulator configuration for the medium module
type Medium struct {
	Bands map[string]Band // Frequency bands in simulation
}
