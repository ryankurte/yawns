package config

// Channels defines channel information for the medium
type Channels struct {
	// Number of channels
	Count uint64
	// Channel Spacing in Hz
	Spacing Frequency
}

// Medium defines the simulator configuration for the medium module
type Medium struct {
	// Radio Frequency in Hz
	Frequency Frequency
	// Baud rate in bps
	Baud Baud
	// Packet overhead in bytes
	Overhead int
	// Standard deviation of gaussian fading in dB
	Fading Attenuation
	// Link Budget in dB
	LinkBudget Attenuation
	// Packet Error Rate
	ErrorRate float64
	// Channel information
	Channels Channels
}
