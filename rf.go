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

package rf

import (
	"fmt"
	"log"
	"math"
)

// Frequency type to assist with unit coherence
type Frequency float64

// Distance type to assist with unit coherence
type Distance float64

const (
	//C is the speed of light in air in meters per second
	C = 2.998e+8

	// FresnelObstructionOK is the largest acceptable proportion of fresnel zone impingement
	FresnelObstructionOK = 0.4
	// FresnelObstructionIdeal is the largest ideal proportion of fresnel zone impingement
	FresnelObstructionIdeal = 0.2

	// Frequency helper types

	Hz  Frequency = 1
	KHz           = Hz * 1000
	MHz           = KHz * 1000
	GHz           = MHz * 1000

	// Distance helper types

	M  Distance = 1
	Km          = M * 1000

	// R is the (average) radius of the earth
	R = 6.371e6

	// Pi for the formulaic use
	π = math.Pi
)

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

// CalculateFoliageLoss calculates path loss in dB due to foliage based on the Weissberger model
// https://en.wikipedia.org/wiki/Weissberger%27s_model
func CalculateFoliageLoss(freq Frequency, depth Distance) (float64, error) {
	if freq < 230*MHz || freq > 95*GHz {
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

// CalculateRaleighFading calculates Raleigh fading
// https://en.wikipedia.org/wiki/Rayleigh_fading
func CalculateRaleighFading(freq Frequency) (float64, error) {
	log.Panicf("Raleigh fading not yet implemented")
	return 0.0, nil
}

// CalculateRicanFading calculates Rican fading
// https://en.wikipedia.org/wiki/Rician_fading
func CalculateRicanFading(freq Frequency) (float64, error) {
	log.Panicf("Rican fading not yet implemented")
	return 0.0, nil
}

// CalculateWeibullFading calculates Weibull fading
// https://en.wikipedia.org/wiki/Weibull_fading
func CalculateWeibullFading(freq Frequency) (float64, error) {
	log.Panicf("Weibull fading not yet implemented")
	return 0.0, nil
}
