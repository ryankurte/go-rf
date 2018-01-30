package rf

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

const allowedError = 0.002

func CheckFloat(actual, expected float64) error {
	if err := math.Abs(actual-expected) / math.Abs(actual); err > allowedError {
		return fmt.Errorf("Actual: %f Expected: %f", actual, expected)
	}
	return nil
}

func TestRFUtils(t *testing.T) {

	t.Run("Can convert from dBm to mW", func(t *testing.T) {

		mw := DecibelMilliVoltToMilliWatt(0.0)
		assert.InDelta(t, 1.0, mw, allowedError)

		mw = DecibelMilliVoltToMilliWatt(10.0)
		assert.InDelta(t, 10.0, mw, allowedError)

		mw = DecibelMilliVoltToMilliWatt(-20.0)
		assert.InDelta(t, 0.01, mw, allowedError)
	})

	t.Run("Can convert from mW to dBm", func(t *testing.T) {

		dbm := MilliWattToDecibelMilliVolt(1.0)
		assert.InDelta(t, 0.0, dbm, allowedError)

		dbm = MilliWattToDecibelMilliVolt(10.0)
		assert.InDelta(t, 10.0, dbm, allowedError)

		dbm = MilliWattToDecibelMilliVolt(0.01)
		assert.InDelta(t, -20, dbm, allowedError)
	})

	t.Run("Can calculate free space attenuation", func(t *testing.T) {

		// Test against precalculated results

		dBLoss := CalculateFreeSpacePathLoss(2.4*GHz, 1e+0)
		assert.InDelta(t, 40.05, float64(dBLoss), allowedError)

		dBLoss = CalculateFreeSpacePathLoss(2.4*GHz, 1e+3)
		assert.InDelta(t, 100.05, float64(dBLoss), allowedError)

		dBLoss = CalculateFreeSpacePathLoss(2.4*GHz, 1e+6)
		assert.InDelta(t, 160.05, float64(dBLoss), allowedError)

		dBLoss = CalculateFreeSpacePathLoss(433*MHz, 1e+3)
		assert.InDelta(t, 85.178, float64(dBLoss), allowedError)

		dBLoss = CalculateFreeSpacePathLoss(433*MHz, 1e+6)
		assert.InDelta(t, 145.178, float64(dBLoss), allowedError)
	})

	t.Run("Can calculate the distance between two lat/lon locations", func(t *testing.T) {
		lat1, lon1 := -36.8485, 174.7633
		lat2, lon2 := -41.2865, 174.7762

		d := CalculateDistance(lat1, lon1, lat2, lon2, R)
		assert.InDelta(t, 493.4e+3, float64(d), 1e+3)
	})

	t.Run("Can calculate fresnel points", func(t *testing.T) {

		// Magic Numbers from: http://www.wirelessconnections.net/calcs/FresnelZone.asp

		zone, err := FresnelFirstZoneMax(2.4*GHz, 10e+3)
		assert.Nil(t, err)
		assert.InDelta(t, 17.671776, zone, allowedError)

		zone, err = FresnelFirstZoneMax(2.4*GHz, 100e+3)
		assert.Nil(t, err)
		assert.InDelta(t, 55.883, zone, allowedError)
	})

	t.Run("Can calculate Fresnel-Kirchoff diffraction parameter", func(t *testing.T) {
		f, d1, d2, h := 900*MHz, 8*Km, 12*Km, -0.334*M

		v, err := CalculateFresnelKirckoffDiffractionParam(f, d1, d2, h)
		assert.Nil(t, err)
		assert.InDelta(t, -0.012, v, 0.0002)
	})

	t.Run("Can calculate Fresnel-Kirchoff loss approximate", func(t *testing.T) {
		v := -0.012

		loss, err := CalculateFresnelKirchoffLossApprox(v)
		assert.Nil(t, err)
		assert.InDelta(t, 5.93, float64(loss), allowedError)
	})

	t.Run("Normalises terrain paths against slope", func(t *testing.T) {
		tests := []struct {
			name         string
			p1, p2, d, l float64
			t            []float64
			x            []float64
			y            []float64
		}{
			{
				"Slope up from L->R",
				0.0, 3.0, 4.0, 5.0,
				[]float64{0.0, 0.0, 0.0},
				[]float64{0.0, 1.6, 3.2},
				[]float64{0.0, -1.2, -2.4},
			}, {
				"Slope down from L->R",
				3.0, 0.0, 4.0, 5.0,
				[]float64{0.0, 0.0, 0.0},
				[]float64{1.8, 3.4, 5.0},
				[]float64{-2.4, -1.2, 0.0},
			}, {
				"Slope up from L->R with positive offset",
				1.0, 4.0, 4.0, 5.0,
				[]float64{0.0, 0.0, 0.0},
				[]float64{-0.6, 1.0, 2.6},
				[]float64{-0.8, -2.0, -3.2},
			}, {
				"Slope up from L->R with positive terrain",
				0.0, 3.0, 4.0, 5.0,
				[]float64{1.0, 1.0, 1.0},
				[]float64{0.6, 2.2, 3.8},
				[]float64{0.8, -0.4, -1.6},
			}, {
				"Slope up from L->R with positive offset and terrain",
				1.0, 4.0, 4.0, 5.0,
				[]float64{1.0, 1.0, 1.0},
				[]float64{0.0, 1.6, 3.2},
				[]float64{0.0, -1.2, -2.4},
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				x, y, l := TerrainToPathXY(test.p1, test.p2, Distance(test.d), test.t)

				assert.InDelta(t, test.l, l, allowedError)

				for i := range test.t {
					assert.InDelta(t, test.x[i], x[i], allowedError, "index %d", i)
					assert.InDelta(t, test.y[i], y[i], allowedError, "index %d", i)
				}
			})
		}
	})

	t.Run("Bullington method (Figure 12) angle calculation", func(t *testing.T) {
		tests := []struct {
			name   string
			x, y   []float64
			d      float64
			θ1, θ2 float64
		}{
			{
				"Zero impingement",
				[]float64{0.0, 2.5, 5.0},
				[]float64{0.0, 0.0, 0.0},
				5, 0.0, 0.0,
			}, {
				"Positive center impingement",
				[]float64{0.0, 2.5, 5.0},
				[]float64{0.0, 1.5, 0.0},
				5, 0.54, 0.54,
			}, {
				"Negative center impingement",
				[]float64{0.0, 2.5, 5.0},
				[]float64{0.0, -1.5, 0.0},
				5, -0.54, -0.54,
			}, {
				"Positive impinging pair",
				[]float64{0.0, 1.0, 2.0, 3.0, 4.0},
				[]float64{0.0, 1.0, 0.0, 1.0, 0.0},
				4, math.Pi / 4, math.Pi / 4,
			}, {
				"Negative impinging pair",
				[]float64{0.0, 1.0, 2.0, 3.0, 4.0},
				[]float64{0.0, -1.0, -4.0, -1.0, 0.0},
				4, math.Atan2(-1, 3), math.Atan2(-1, 3),
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				θ1, θ2 := findBullingtonFigure12Angles(test.x, test.y, test.d)
				assert.InDelta(t, test.θ1, θ1, allowedError)
				assert.InDelta(t, test.θ2, θ2, allowedError)
			})
		}
	})

	t.Run("Bullington method (figure 12) angles to distance", func(t *testing.T) {
		tests := []struct {
			name      string
			θ1, θ2, l float64
			h, d      float64
		}{
			{
				"Positive and equal angles",
				math.Pi / 4, math.Pi / 4, 10,
				5, 5,
			}, {
				"Negative and equal angles",
				-math.Pi / 4, -math.Pi / 4, 10,
				-5, 5,
			}, {
				"Positive and left skewed",
				math.Pi / 3, math.Pi / 6, 10,
				10 * math.Sin(math.Pi/3) * math.Sin(math.Pi/6), 10 * math.Sin(math.Pi/6) * math.Cos(math.Pi/3),
			}, {
				"Positive and right skewed",
				math.Pi / 6, math.Pi / 3, 10,
				10 * math.Sin(math.Pi/6) * math.Sin(math.Pi/3), 10 * math.Sin(math.Pi/3) * math.Cos(math.Pi/6),
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				d, h := solveBullingtonFigureTwelveDist(test.θ1, test.θ2, test.l)
				assert.InDelta(t, test.h, h, allowedError)
				assert.InDelta(t, test.d, d, allowedError)
			})
		}
	})

	t.Run("Reverts dist/height to x/y", func(t *testing.T) {
		tests := []struct {
			name         string
			p1, p2, d    float64
			dist, height float64
			x, y         float64
		}{
			{
				"Zero path height",
				0.0, 0.0, 4.0,
				2.0, 1.0,
				2.0, 1.0,
			}, {
				"Constant path height",
				1.0, 1.0, 4.0,
				2.0, 1.0,
				2.0, 2.0,
			}, {
				"Angled path with no offset ",
				0.0, 3.0, 4.0,
				2.5, 0.0,
				2.0, 1.5,
			}, {
				"Angled path with offset ",
				0.0, 3.0, 4.0,
				2.5, 1.0,
				1.4, 2.3,
			}, {
				"Reverse angled path with offset ",
				3.0, 0.0, 4.0,
				2.5, 1.0,
				2.6, 2.3,
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				x, y := UnNormalisePoint(test.p1, test.p2, Distance(test.d), test.dist, test.height)
				assert.InDelta(t, test.x, x, allowedError, "X")
				assert.InDelta(t, test.y, y, allowedError, "Y")
			})
		}
	})

	t.Run("Computes maximum fresnel zone impingement over terrain", func(t *testing.T) {
		tests := []struct {
			name   string
			p1, p2 float64
			d      Distance
			f      Frequency
			t      []float64
			i, p   float64
		}{
			{
				"No impingement",
				0.0, 0.0, 50.0 * M, 433 * MHz,
				[]float64{-100.0, -100.0, -100.0, -100.0, -100.0},
				0.0, 25.0,
			}, {
				"50% impingement",
				0.0, 0.0, 50.0 * M, 433 * MHz,
				[]float64{-100.0, -100.0, 0.0, -100.0, -100.0},
				0.5, 25.0,
			}, {
				"100% impingement",
				0.0, 0.0, 50.0 * M, 433 * MHz,
				[]float64{-100.0, -100.0, 2.94, -100.0, -100.0},
				1.0, 25.0,
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				i, p, err := FresnelImpingementMax(test.p1, test.p2, test.d, test.f, test.t)
				assert.Nil(t, err)

				assert.InDelta(t, float64(test.i), float64(i), allowedError)
				assert.InDelta(t, float64(test.p), float64(p), allowedError)

			})
		}
	})

}
