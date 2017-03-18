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
	"fmt"
	"math"
)

const (
	//C is the speed of light in air in meters per second
	C                       = 2.998e+8
	FresnelObstructionOK    = 0.4
	FresnelObstructionIdeal = 0.2
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
// Note that distances must be much greater than wavelengths
// https://en.wikipedia.org/wiki/Fresnel_zone#Fresnel_zone_clearance

const FresnelMinDistanceFreqRadio = 0.1

// FresnelPoint calculates the fresnel zone radius d for a given wavelength
// and order at a point P between endpoints
func FresnelPoint(d1, d2, freq float64, order int64) (float64, error) {
	wavelength := FrequencyToWavelength(freq)

	if ((d1 * FresnelMinDistanceFreqRadio) < wavelength) || ((d2 * FresnelMinDistanceFreqRadio) < wavelength) {
		return 0, fmt.Errorf("Fresnel calculation valid only for distances >> wavelength (d1: %.2fm d2: %.2fm wavelength %.2fm)", d1, d2, wavelength)
	}

	return math.Sqrt((float64(order) * wavelength * d1 * d2) / (d1 + d2)), nil
}

// FresnelFirstZoneMax calculates the maximum fresnel zone radius for a given frequency
func FresnelFirstZoneMax(freq, dist float64) (float64, error) {

	wavelength := FrequencyToWavelength(freq)
	if (dist * FresnelMinDistanceFreqRadio) < wavelength {
		return 0, fmt.Errorf("Fresnel calculation valid only for distance >> wavelength (distance: %.2fm wavelength %.2fm)", dist, wavelength)
	}

	return 0.5 * math.Sqrt((C * dist / 1000 / freq)), nil
}
