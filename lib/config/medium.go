package config

// Channels defines channel information for the medium
type Channels struct {
	// Number of channels
	Count uint64
	// Channel Spacing in Hz
	Spacing float64
}

// Medium defines the simulator configuration for the medium module
type Medium struct {
	// Radio Frequency in Hz
	Frequency float64
	// Standard deviation of gaussian fading in dB
	Fading float64
	// Link Budget in dB
	LinkBudget float64
	// Packet Error Rate
	ErrorRate float64
	// Channel information
	Channels Channels
}
