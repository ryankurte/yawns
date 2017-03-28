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

// Frequency type to assist with unit coherence
type Frequency float64

// Distance type to assist with unit coherence
type Distance float64

const (
	//C is the speed of light in air in meters per second
	C = 2.998e+8

	// FresnelObstructionOK is the largets acceptable proportion of fresnel zone impingement
	FresnelObstructionOK = 0.4
	// FresnelObstructionIdeal is the largest ideal proportion of fresnel zone impingement
	FresnelObstructionIdeal = 0.2

	Hz  Frequency = 1
	KHz           = Hz * 1000
	MHz           = KHz * 1000
	GHz           = MHz * 1000

	M  Distance = 1
	Km          = M * 1000

	// R is the (average) radius of the earth
	R = 6.371e6

	π = math.Pi
)

// Basic RF calculations

// FrequencyToWavelength calculates a wavelength from a frequency
func FrequencyToWavelength(freq Frequency) float64 {
	return float64(C / freq)
}

// WavelengthToFrequency calculates a frequency from a wavelength
func WavelengthToFrequency(wavelength float64) Frequency {
	return Frequency(C / wavelength)
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
func FreeSpaceAttenuation(freq Frequency, distance Distance) float64 {
	return math.Pow((4 * math.Pi * float64(distance) * float64(freq) / C), 2)
}

// FreeSpaceAttenuationDB calculates the Free Space Path Loss for a given frequency and distance in Decibels
func FreeSpaceAttenuationDB(freq Frequency, distance Distance) float64 {
	return 20 * math.Log10((4 * math.Pi * float64(distance) * float64(freq) / C))
}

// Freznel zone calculations
// Note that distances must be much greater than wavelengths
// https://en.wikipedia.org/wiki/Fresnel_zone#Fresnel_zone_clearance

// FresnelMinDistanceWavelengthRadio is the minimum ratio of distance:wavelength for viable calculations
// This is used as a programattic sanity check for distance >> wavelength
const FresnelMinDistanceWavelengthRadio = 0.1

// FresnelPoint calculates the fresnel zone radius d for a given wavelength
// and order at a point P between endpoints
func FresnelPoint(d1, d2 Distance, freq Frequency, order int64) (float64, error) {
	wavelength := FrequencyToWavelength(freq)

	if ((float64(d1) * FresnelMinDistanceWavelengthRadio) < wavelength) || ((float64(d2) * FresnelMinDistanceWavelengthRadio) < wavelength) {
		return 0, fmt.Errorf("Fresnel calculation valid only for distances >> wavelength (d1: %.2fm d2: %.2fm wavelength %.2fm)", d1, d2, wavelength)
	}

	return math.Sqrt((float64(order) * wavelength * float64(d1) * float64(d2)) / (float64(d1) + float64(d2))), nil
}

// FresnelFirstZoneMax calculates the maximum fresnel zone radius for a given frequency
func FresnelFirstZoneMax(freq Frequency, dist Distance) (float64, error) {

	wavelength := FrequencyToWavelength(freq)
	if (float64(dist) * FresnelMinDistanceWavelengthRadio) < wavelength {
		return 0, fmt.Errorf("Fresnel calculation valid only for distance >> wavelength (distance: %.2fm wavelength %.2fm)", dist, wavelength)
	}

	return 0.5 * math.Sqrt((C * float64(dist) / float64(freq))), nil
}

// CalculateDistance calculates the distance between two latitude and longitudes
// Using the haversine (flat earth) formula
// See: http://www.movable-type.co.uk/scripts/latlong.html
func CalculateDistance(lat1, lng1, lat2, lng2, radius float64) float64 {

	φ1, λ1 := lat1/180*π, lng1/180*π
	φ2, λ2 := lat2/180*π, lng2/180*π
	Δφ, Δλ := math.Abs(φ2-φ1), math.Abs(λ2-λ1)

	a := math.Pow(math.Sin(Δφ/2), 2) + math.Cos(φ1)*math.Cos(φ2)*math.Pow(math.Sin(Δλ/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	d := radius * c

	return d
}

// CalculateDistanceLOS calculates the approximate Line of Sight distance between two lat/lon/alt points
// This achieves this by wrapping the haversine formula with a flat-earth approximation for height
// difference. This will be very inaccurate with larger distances.
// TODO: surely there is a better (ie. not written by me) algorithm for this
func CalculateDistanceLOS(lat1, lng1, alt1, lat2, lng2, alt2 float64) float64 {

	// Calculate average and delta heights (wrt. earth radius)
	h := R + (alt1+alt2)/2
	Δh := math.Abs(alt2 - alt1)

	// Compute distance at average of altitudes
	d := CalculateDistance(lat1, lng1, lat2, lng2, h)

	// Apply transformation for altitude difference
	los := math.Sqrt(math.Pow(d, 2) + math.Pow(Δh, 2))

	return los
}

// CalculateFoliageLoss calculates path loss in dB due to foliage based on the Weissberger model
// https://en.wikipedia.org/wiki/Weissberger%27s_model
func CalculateFoliageLoss(freq Frequency, depth Distance) (float64, error) {
	if freq < 230e6 || freq > 95e9 {
		return 0, fmt.Errorf("Frequency %.2f is not between 230MHz and 95GHz as required by the Weissberger model", freq)
	}

	if depth > 400 || depth < 0 {
		return 0, fmt.Errorf("Depth %.2f is not between 0 and 400m as required by the Weissberger model", depth)
	}

	fading := 0.0
	if depth > 00.0 && depth <= 14.0 {
		fading = 0.45 * math.Pow(float64(freq), 0.284) * float64(depth)
	} else if depth > 14.0 && depth <= 400.0 {
		fading = 1.33 * math.Pow(float64(freq), 0.284) * math.Pow(float64(depth), 0.588)
	}

	return fading, nil
}

// https://en.wikipedia.org/wiki/Rayleigh_fading
func CalculateRaleighFading(freq Frequency) (float64, error) {

	return 0.0, nil
}

// https://en.wikipedia.org/wiki/Rician_fading
func CalculateRicanFading(freq Frequency) (float64, error) {

	return 0.0, nil
}

// https://en.wikipedia.org/wiki/Weibull_fading
func CalculateWeibullFading(freq Frequency) (float64, error) {

	return 0.0, nil
}
