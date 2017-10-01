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
		p1, p2, d, tr := 0.0, 3.0, Distance(4.0), []float64{0.0, 0.0, 0.0}
		x, y, d2 := TerrainToPathXY(p1, p2, d, tr)

		fmt.Printf("X: %+v Y: %+v\n", x, y)

		assert.InDelta(t, 5.0, d2, allowedError)

		assert.InDelta(t, 0.0, x[0], allowedError)
		assert.InDelta(t, 0.0, y[0], allowedError)

		assert.InDelta(t, 1.6, x[1], allowedError)
		assert.InDelta(t, 1.2, y[1], allowedError)

		assert.InDelta(t, 3.2, x[2], allowedError)
		assert.InDelta(t, 2.4, y[2], allowedError)
	})

	t.Run("Normalises terrain paths against slope2", func(t *testing.T) {
		p1, p2, d, tr := 3.0, 0.0, Distance(4.0), []float64{0.0, 0.0, 0.0}
		x, y, d2 := TerrainToPathXY(p1, p2, d, tr)

		fmt.Printf("X: %+v Y: %+v\n", x, y)

		assert.InDelta(t, 5.0, d2, allowedError)

		assert.InDelta(t, 1.8, x[0], allowedError)
		assert.InDelta(t, 2.4, y[0], allowedError)

		assert.InDelta(t, 3.4, x[1], allowedError)
		assert.InDelta(t, 1.2, y[1], allowedError)

		assert.InDelta(t, 5.0, x[2], allowedError)
		assert.InDelta(t, 0.0, y[2], allowedError)
	})

	t.Run("Normalises terrain paths for Fresnel-Kirchoff method", func(t *testing.T) {
		h, d := TerrainToFresnelKirchoff(1.0, 1.0, 2.0, []float64{0.0, 0.0, 0.0, 0.0, 0.0})
		assert.InDelta(t, -1.0, h, allowedError)
		assert.InDelta(t, 1.0, d, allowedError)

		h, d = TerrainToFresnelKirchoff(1.0, 1.0, 2.0, []float64{0.0, 0.5, 0.0, 0.0, 0.0})
		assert.InDelta(t, -0.5, h, allowedError)
		assert.InDelta(t, 0.5, d, allowedError)

		h, d = TerrainToFresnelKirchoff(1.0, 1.0, 2.0, []float64{0.0, 1.0, 1.0, 0.0, 0.0})
		assert.InDelta(t, 0.0, h, allowedError)
		assert.InDelta(t, 0.75, d, allowedError)

		h, d = TerrainToFresnelKirchoff(1.0, 1.0, 2.0, []float64{0.0, 0.0, 1.5, 1.5, 0.0})
		assert.InDelta(t, 0.5, h, allowedError)
		assert.InDelta(t, 1.25, d, allowedError)

		h, d = TerrainToFresnelKirchoff(2.0, 1.0, 2.0, []float64{0.0, 0.0, 1.0, 0.0, 0.0})
		assert.InDelta(t, -0.4436, h, allowedError)
		assert.InDelta(t, 1.0, d, allowedError)

	})

}
