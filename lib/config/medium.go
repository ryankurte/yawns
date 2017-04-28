package config

// Channels defines channel information for the medium
type Channels struct {
	Count   uint64    // Number of channels
	Spacing Frequency // Channel Spacing in Hz
}

// Medium defines the simulator configuration for the medium module
type Medium struct {
	Frequency       Frequency   // Radio Frequency in Hz
	Baud            Baud        // Baud rate in bps
	Overhead        int         // Packet overhead in bytes
	RandomDeviation Attenuation // Standard deviation of gaussian fading in dB
	LinkBudget      Attenuation // Link Budget in dB
	ErrorRate       float64     // Packet Error Rate
	Channels        Channels    // Channel information
}
