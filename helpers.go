package rf

import (
	"math"
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

// Distance and Radius calculations

// CalculateDistance calculates the distance between two latitude and longitudes
// Using the haversine (flat earth) formula
// See: http://www.movable-type.co.uk/scripts/latlong.html
func CalculateDistance(lat1, lng1, lat2, lng2, radius float64) Distance {

	φ1, λ1 := lat1/180*π, lng1/180*π
	φ2, λ2 := lat2/180*π, lng2/180*π
	Δφ, Δλ := math.Abs(φ2-φ1), math.Abs(λ2-λ1)

	a := math.Pow(math.Sin(Δφ/2), 2) + math.Cos(φ1)*math.Cos(φ2)*math.Pow(math.Sin(Δλ/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	d := radius * c

	return Distance(d)
}

// CalculateDistanceLOS calculates the approximate Line of Sight distance between two lat/lon/alt points
// This achieves this by wrapping the haversine formula with a flat-earth approximation for height
// difference. This will be very inaccurate with larger distances.
// TODO: surely there is a better (ie. not written by me) algorithm for this
func CalculateDistanceLOS(lat1, lng1, alt1, lat2, lng2, alt2 float64) Distance {

	// Calculate average and delta heights (wrt. earth radius)
	h := R + (alt1+alt2)/2
	Δh := math.Abs(alt2 - alt1)

	// Compute distance at average of altitudes
	d := CalculateDistance(lat1, lng1, lat2, lng2, h)

	// Apply transformation for altitude difference
	los := math.Sqrt(math.Pow(float64(d), 2) + math.Pow(Δh, 2))

	return Distance(los)
}
