package config

import (
	"github.com/ryankurte/ons/lib/types"
)

// Channels defines channel information for the medium
type Channels struct {
	Count   uint64          // Number of channels
	Spacing types.Frequency // Channel Spacing in Hz
}

// Band is a simulated frequency band
type Band struct {
	Frequency       types.Frequency   // Radio Frequency in Hz
	Baud            types.Baud        // Baud rate in bps
	PacketOverhead  types.SizeBytes   // Packet overhead in bytes
	RandomDeviation types.Attenuation // Standard deviation of gaussian fading in dB
	LinkBudget      types.Attenuation // Link Budget in dB
	ErrorRate       float64           // Packet Error Rate
	Channels        Channels          // Channel information
}

// Medium defines the simulator configuration for the medium module
type Medium struct {
	Bands map[string]Band // Frequency bands in simulation
}
