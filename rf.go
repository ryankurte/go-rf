/*
 * Radio Frequency calculations
 *
 * Note that attenuation is a field quantity and thus Decibels (dB) are defined as 20log10
 * rather than the common 10log10 used for power measurements
 * See: https://en.wikipedia.org/wiki/Decibel#Field_quantities_and_root-power_quantities
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

// Frequency type (Hz) to assist with unit coherence
type Frequency float64

// Wavelength type (m) to assist with unit coherence
type Wavelength float64

// Distance type (m) to assist with unit coherence
type Distance float64

// Attenuation type (dB) to assist with unit coherence
type Attenuation float64

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

	// Attenuation helpers

	DB Attenuation = 0

	// R is the (average) radius of the earth
	R = 6.371e6

	// Pi for use in formulae
	π = math.Pi
)

// Free Space Path Loss (FSPL) calculations
// https://en.wikipedia.org/wiki/Free-space_path_loss#Free-space_path_loss_formula

// CalculateFreeSpacePathLoss calculates the Free Space Path Loss in Decibels for a given frequency and distance
func CalculateFreeSpacePathLoss(freq Frequency, distance Distance) Attenuation {
	fading := 20 * math.Log10((4 * math.Pi * float64(distance) * float64(freq) / C))
	return Attenuation(fading)
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

	if ((float64(d1) * FresnelMinDistanceWavelengthRadio) < float64(wavelength)) || ((float64(d2) * FresnelMinDistanceWavelengthRadio) < float64(wavelength)) {
		return 0, fmt.Errorf("Fresnel calculation valid only for distances >> wavelength (d1: %.2fm d2: %.2fm wavelength %.2fm)", d1, d2, wavelength)
	}

	return math.Sqrt((float64(order) * float64(wavelength) * float64(d1) * float64(d2)) / (float64(d1) + float64(d2))), nil
}

// FresnelFirstZoneMax calculates the maximum fresnel zone radius for a given frequency
func FresnelFirstZoneMax(freq Frequency, dist Distance) (float64, error) {

	wavelength := FrequencyToWavelength(freq)
	if (float64(dist) * FresnelMinDistanceWavelengthRadio) < float64(wavelength) {
		return 0, fmt.Errorf("Fresnel calculation valid only for distance >> wavelength (distance: %.2fm wavelength %.2fm)", dist, wavelength)
	}

	return 0.5 * math.Sqrt((C * float64(dist) / float64(freq))), nil
}

// CalculateFresnelKirckoffDiffractionParam Calculates the Fresnel-Kirchoff Diffraction parameter
// https://en.wikipedia.org/wiki/Kirchhoff%27s_diffraction_formula
// https://s.campbellsci.com/documents/au/technical-papers/line-of-sight-obstruction.pdf
// d1 and d2 are the distances between the "knife edge" impingement and the transmitter/receiver
// h is the impingement, where -ve is below LoS and +ve is above LoS
func CalculateFresnelKirckoffDiffractionParam(freq Frequency, d1, d2, h Distance) (float64, error) {
	wavelength := FrequencyToWavelength(freq)
	v := float64(h) * math.Sqrt((2*float64(d1+d2))/(float64(wavelength)*float64(d1*d2)))
	return v, nil
}

// CalculateFresnelKirchoffLossApprox Calculates approximate loss due to diffraction using
// the Fresnel-Kirchoff Diffraction parameter. This approximate is valid for values >= -0.7
// https://s.campbellsci.com/documents/au/technical-papers/line-of-sight-obstruction.pdf
func CalculateFresnelKirchoffLossApprox(v float64) (Attenuation, error) {
	if !(v >= -0.7) {
		return 0.0, fmt.Errorf("Fresnel-Kirchoff loss approximation only valid for v >= -0.7")
	}
	loss := 6.9 + 20*math.Log10(math.Sqrt(math.Pow(v-0.1, 2)+1)+v-0.1)
	return Attenuation(loss), nil
}

const (
	WeissbergerMinFreq = 230 * MHz
	WeissbergerMaxFreq = 95 * GHz
	WeissbergerMinDist = 0 * M
	WeissbergerMaxDist = 400 * M
)

// CalculateFoliageLoss calculates path loss in dB due to foliage based on the Weissberger model
// https://en.wikipedia.org/wiki/Weissberger%27s_model
func CalculateFoliageLoss(freq Frequency, depth Distance) (Attenuation, error) {
	if freq < WeissbergerMinFreq || freq > WeissbergerMaxFreq {
		return 0, fmt.Errorf("Frequency %.2f is not between 230MHz and 95GHz as required by the Weissberger model", freq)
	}

	if depth < WeissbergerMinDist || WeissbergerMaxDist > 0 {
		return 0, fmt.Errorf("Depth %.2f is not between 0 and 400m as required by the Weissberger model", depth)
	}

	fading := 0.0
	if depth > 00.0 && depth <= 14.0 {
		fading = 0.45 * math.Pow(float64(freq), 0.284) * float64(depth)
	} else if depth > 14.0 && depth <= 400.0 {
		fading = 1.33 * math.Pow(float64(freq), 0.284) * math.Pow(float64(depth), 0.588)
	}

	return Attenuation(fading), nil
}

// CalculateRaleighFading calculates Raleigh fading
// https://en.wikipedia.org/wiki/Rayleigh_fading
func CalculateRaleighFading(freq Frequency) (Attenuation, error) {
	log.Panicf("Raleigh fading not yet implemented")
	return 0.0, nil
}

// CalculateRicanFading calculates Rican fading
// https://en.wikipedia.org/wiki/Rician_fading
func CalculateRicanFading(freq Frequency) (Attenuation, error) {
	log.Panicf("Rican fading not yet implemented")
	return 0.0, nil
}

// CalculateWeibullFading calculates Weibull fading
// https://en.wikipedia.org/wiki/Weibull_fading
func CalculateWeibullFading(freq Frequency) (Attenuation, error) {
	log.Panicf("Weibull fading not yet implemented")
	return 0.0, nil
}
