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

	// R is the (average) radius of the earth
	R = 6.371e6

	// Pi for use in formulae
	π = math.Pi
)

// Frequency helper types
const (
	Hz  Frequency = 1.0
	KHz           = Hz * 1000
	MHz           = KHz * 1000
	GHz           = MHz * 1000
)

// Distance helper types
const (
	M  Distance = 1.0
	Km          = M * 1000
)

// Attenuation helpers
const (
	DB Attenuation = 0
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
// d1 and d2 are the distances between the "knife edge" impingement and the transmitter/receiver
// h is the impingement, where -ve is below Line of Sight (LoS) and +ve is above LoS
// https://en.wikipedia.org/wiki/Kirchhoff%27s_diffraction_formula
// https://s.campbellsci.com/documents/au/technical-papers/line-of-sight-obstruction.pdf
func CalculateFresnelKirckoffDiffractionParam(freq Frequency, d1, d2, h Distance) (v float64, err error) {
	wavelength := FrequencyToWavelength(freq)
	v = float64(h) * math.Sqrt((2*float64(d1+d2))/(float64(wavelength)*float64(d1*d2)))
	return v, err
}

// CalculateFresnelKirchoffLossApprox Calculates approximate loss due to diffraction using
// the Fresnel-Kirchoff Diffraction parameter. This approximate is valid for values >= -0.7
// https://s.campbellsci.com/documents/au/technical-papers/line-of-sight-obstruction.pdf
func CalculateFresnelKirchoffLossApprox(v float64) (Attenuation, error) {
	if !(v >= -0.7) {
		return 0.0, fmt.Errorf("Fresnel-Kirchoff loss approximation only valid for v >= -0.7 (v: %.6f)", v)
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

	if depth < WeissbergerMinDist || depth > WeissbergerMaxDist {
		return 0, fmt.Errorf("Depth %.2f is not between 0 and 400m as required by the Weissberger model", depth)
	}

	fading := 0.0
	if depth > 0.0 && depth <= 14.0 {
		fading = 0.45 * math.Pow(float64(freq/GHz), 0.284) * float64(depth)
	} else if depth > 14.0 && depth <= 400.0 {
		fading = 1.33 * math.Pow(float64(freq/GHz), 0.284) * math.Pow(float64(depth), 0.588)
	} else {
		return 0, fmt.Errorf("Depth %.2f is not between 0 and 400m as required by the Weissberger model", depth)
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

// BullingtonFigure12Method implements the Bullington Figure 12 (intersecting horizons) method to approximate
// height and distance for use in the Fresnell-Kirchoff path loss approximation.
// Note that this implementation is not accurate for most negative (below LOS) impingements
// See: https://hams.soe.ucsc.edu/sites/default/files/Bullington%20VTS%201977.pdf
func BullingtonFigure12Method(p1, p2 float64, d Distance, terrain []float64) (d1, d2, height float64) {
	x, y, l := TerrainToPathXY(p1, p2, d, terrain)

	θ1, θ2 := findBullingtonFigure12Angles(x, y, l)

	d1, height = solveBullingtonFigureTwelveDist(θ1, θ2, l)
	d2 = l - d1

	return d1, d2, height
}

func findBullingtonFigure12Angles(x, y []float64, d float64) (θ1, θ2 float64) {
	// Find minimum angles
	maxθ1, maxθ2 := -math.Pi/2, -math.Pi/2

	for i := 1; i < len(x)-1; i++ {
		θ1 := math.Atan2(y[i], x[i])
		θ2 := math.Atan2(y[i], d-x[i])

		if θ1 > maxθ1 {
			maxθ1 = θ1
		}

		if θ2 > maxθ2 {
			maxθ2 = θ2
		}
	}

	return maxθ1, maxθ2
}

func solveBullingtonFigureTwelveDist(θb, θc, l float64) (dist, height float64) {
	θa := math.Pi - θb - θc

	r := l / math.Sin(θa)

	C := r * math.Sin(θc)
	height = math.Sin(θb) * C
	dist = math.Cos(θb) * C

	return dist, height
}

// FresnelImpingementMax computes the maximum first fresnel zone impingement due to terrain between two points
func FresnelImpingementMax(p1, p2 float64, d Distance, f Frequency, terrain []float64) (maxImpingement float64, point Distance) {
	x, y, l := TerrainToPathXY(p1, p2, d, terrain)

	maxImpingement, point = 0.0, Distance(l/2)

	for i := 1; i < len(x)-1; i++ {
		d1 := Distance(x[i])
		d2 := Distance(l) - d1

		// Calculate size of fresnel zone
		fresnelZone, err := FresnelPoint(d1, d2, f, 1)
		if err != nil {
			// Skip invalid points (where wavelength is not << d1 or d2)
			continue
		}

		// Calculate impingement
		impingement := 0.0
		if y[i] > fresnelZone/2 {
			impingement = 1.0
		} else if y[i] < -fresnelZone/2 {
			impingement = 0.0
		} else {
			impingement = (y[i] + fresnelZone/2) / fresnelZone
		}

		// Record max
		if impingement > maxImpingement {
			maxImpingement = impingement
			point = d1
		}
	}

	return maxImpingement, point
}
