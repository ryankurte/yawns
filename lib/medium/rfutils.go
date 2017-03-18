/*
 * Radio Frequency calculations
 *
 *
 * More Reading:
 * https://en.wikipedia.org/wiki/Path_loss
 * https://en.wikipedia.org/wiki/Friis_transmission_equation
 * https://en.wikipedia.org/wiki/Hata_model_for_urban_areas
 * https://en.wikipedia.org/wiki/Rayleigh_fading
 * https://en.wikipedia.org/wiki/Rician_fading
 *
 * Copyright 2017 Ryan Kurte
 */

package medium

import (
	"math"
)

const (
	//C is the speed of light in air in meters per second
	C = 2.99792458e+8
)

// Basic RF calculations

// FrequencyToWavelength calculates a wavelength from a frequency
func FrequencyToWavelength(freq float64) float64 {
	return C / freq
}

// WavelengthToFrequency calculates a frequency from a wavelength
func WavelengthToFrequency(wavelength float64) float64 {
	return C / wavelength
}

// DecibelMilliVoltToMilliWatt converts dBm to mW
func DecibelMilliVoltToMilliWatt(dbm float64) float64 {
	return math.Pow(10, dbm/10)
}

// MilliWattToDecibelMilliVolt converts mW to dBm
func MilliWattToDecibelMilliVolt(mw float64) float64 {
	return 10 * math.Log10(mw)
}

// Free Space Path Loss (FSPL) calculations
// https://en.wikipedia.org/wiki/Free-space_path_loss#Free-space_path_loss_formula

// FreeSpaceAttenuation calculates the Free Space Path Loss for a given frequency and distance
func FreeSpaceAttenuation(freq, distance float64) float64 {
	return math.Pow((4 * math.Pi * distance * freq / C), 2)
}

// FreeSpaceAttenuationDB calculates the Free Space Path Loss for a given frequency and distance in Decibels
func FreeSpaceAttenuationDB(freq, distance float64) float64 {
	return 20 * math.Log10((4 * math.Pi * distance * freq / C))
}

// Freznel zone calculations
// https://en.wikipedia.org/wiki/Fresnel_zone#Fresnel_zone_clearance

// FresnelPoint calculates the fresnel zone radius d for a given wavelength
// and order at a point P between endpoints
func FresnelPoint(d1, d2, freq float64, order int64) float64 {
	wavelength := FrequencyToWavelength(freq)
	return math.Sqrt((float64(order) * wavelength * d1 * d2) / (d1 + d2))
}

// FresnelMax calculates the maximum fresnel zone radius for a given wavelength and order
func FresnelMax(freq float64, order int64) float64 {
	wavelength := FrequencyToWavelength(freq)
	return math.Sqrt((float64(order) * wavelength) / 2)
}
