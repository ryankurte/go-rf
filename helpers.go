package rf

import (
	"fmt"
	"math"
)

// Basic RF calculations

// FrequencyToWavelength calculates a wavelength from a frequency
func FrequencyToWavelength(freq Frequency) Wavelength {
	return Wavelength(C / freq)
}

// WavelengthToFrequency calculates a frequency from a wavelength
func WavelengthToFrequency(wavelength Wavelength) Frequency {
	return Frequency(C / wavelength)
}

// Power Decibel helpers
// See https://en.wikipedia.org/wiki/Decibel#Power_quantities

// DecibelMilliVoltToMilliWatt converts dBm to mW
// Note that this power decibels (10log10)
func DecibelMilliVoltToMilliWatt(dbm float64) float64 {
	return math.Pow(10, dbm/10)
}

// MilliWattToDecibelMilliVolt converts mW to dBm
// Note that this power decibels (10log10)
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

// FieldDBToAbs Converts field attenuation (20log10) to absolute values
func (a *Attenuation) FieldDBToAbs() float64 {
	return math.Pow(10, float64(*a)/20)
}

// FieldAbsToDB Converts an absolute field attenuation (20log10) to decibels
func FieldAbsToDB(abs float64) Attenuation {
	return Attenuation(20 * math.Log10(abs))
}

// Convert terrain between two points of set heights into distances from the path between those points
func TerrainToPath(p1, p2 float64, d Distance, terrain []float64) (Δd, Δh, θ float64, diffs []float64) {
	height := (p2 - p1)
	θ = math.Sin(height / float64(d))
	dist := math.Cos(θ) * float64(d)

	Δh = height / float64(len(terrain)-1)
	Δd = dist / float64(len(terrain)-1)

	diffs = make([]float64, len(terrain))

	fmt.Printf("height: %.4f dist: %.4f θ: %.4f Δh: %.4f Δd: %.4f\n", height, dist, θ, Δh, Δd)

	for i, v := range terrain {
		h := p1 + float64(i)*Δh
		d := v - h
		nh := math.Cos(θ) * d

		fmt.Printf("Slice %d dist: %.4f height: %.4f terrain: %.4f diff: %.4f normalised: %.4f\n", i, float64(i)*Δd, h, v, d, nh)

		diffs[i] = nh
	}

	return Δd, Δh, θ, diffs
}

// TerrainToPathXY Converts terrain between two points of set heights into distances from the path between those points
func TerrainToPathXY(p1, p2 float64, d Distance, terrain []float64) (x, y []float64, d2 float64) {
	height := (p2 - p1)
	θ := math.Atan2(height, float64(d))

	Δh := height / float64(len(terrain)-1)
	Δd := float64(d) / float64(len(terrain)-1)

	x = make([]float64, len(terrain))
	y = make([]float64, len(terrain))

	fmt.Printf("TXY height: %.4f dist: %.4f θ: %.4f Δh: %.4f Δd: %.4f\n", height, d, θ, Δh, Δd)

	for i, terrainHeight := range terrain {
		referenceHeight := p1 + float64(i)*Δh
		referenceDist := Δd * float64(i)

		verticalClearance := referenceHeight - terrainHeight

		transformedX := math.Sin(θ) * verticalClearance
		transformedY := math.Cos(θ) * verticalClearance

		shiftX := referenceDist / math.Cos(θ)

		x[i], y[i] = shiftX-transformedX, transformedY

		fmt.Printf("Point: %d terrain: %.2f ph: %.2f pd: %.2f clearance: %.2f sx: %.2f tx: %.2f ty: %.2f x: %.2f y: %.2f\n",
			i, terrainHeight, referenceHeight, referenceDist, verticalClearance, shiftX, transformedX, transformedY, x[i], y[i])

	}

	d2 = math.Sqrt(math.Pow(height, 2) + math.Pow(float64(d), 2))

	return x, y, d2
}

func TerrainToFresnelKirchoff(p1, p2 float64, d Distance, terrain []float64) (highestImpingement, distanceToImpingement float64) {
	Δd, _, θ, diffs := TerrainToPath(p1, p2, d, terrain)

	highestIndex := 0
	count := 0
	for i, v := range diffs {
		if v > highestImpingement || i == 0 {
			highestImpingement = v
			highestIndex = i
			count = 0
		} else if v == highestImpingement {
			count++
		}
	}
	distanceToImpingement = (float64(highestIndex) + float64(count)/2) * Δd / math.Cos(θ)

	return highestImpingement, distanceToImpingement
}

func BullingtonMethod(p1, p2 float64, d Distance, terrain []float64) (highestImpingement, distanceToImpingement float64) {
	//x, y := TerrainToPathXY(p1, p2, d, terrain)

	return highestImpingement, distanceToImpingement
}
