package medium

import (
	"fmt"
	"math"
	"testing"
)

const allowedError = 0.001

func CheckFloat(actual, expected float64) error {
	if err := math.Abs(actual-expected) / math.Abs(actual); err > allowedError {
		return fmt.Errorf("Actual: %f Expected: %f", actual, expected)
	}
	return nil
}

func TestRFUtils(t *testing.T) {

	t.Run("Can convert from dBm to mW", func(t *testing.T) {

		mw := DecibelMilliVoltToMilliWatt(0.0)
		err := CheckFloat(mw, 1.0)
		if err != nil {
			t.Error(err)
		}

		mw = DecibelMilliVoltToMilliWatt(10.0)
		err = CheckFloat(mw, 10.0)
		if err != nil {
			t.Error(err)
		}

		mw = DecibelMilliVoltToMilliWatt(-20.0)
		err = CheckFloat(mw, 0.01)
		if err != nil {
			t.Error(err)
		}

	})

	t.Run("Can convert from mW to dBm", func(t *testing.T) {

		dbm := MilliWattToDecibelMilliVolt(1.0)
		err := CheckFloat(dbm, 0.0)
		if err != nil {
			t.Error(err)
		}

		dbm = MilliWattToDecibelMilliVolt(10.0)
		err = CheckFloat(dbm, 10.0)
		if err != nil {
			t.Error(err)
		}

		dbm = MilliWattToDecibelMilliVolt(0.01)
		err = CheckFloat(dbm, -20)
		if err != nil {
			t.Error(err)
		}

	})

	t.Run("Can calculate free space attenuation", func(t *testing.T) {

		// Test against precalculated results

		dBLoss := FreeSpaceAttenuationDB(2.4*GHz, 1e+0)
		err := CheckFloat(dBLoss, 40.02)
		if err != nil {
			t.Error(err)
		}

		dBLoss = FreeSpaceAttenuationDB(2.4*GHz, 1e+3)
		err = CheckFloat(dBLoss, 100.05)
		if err != nil {
			t.Error(err)
		}

		dBLoss = FreeSpaceAttenuationDB(2.4*GHz, 1e+6)
		err = CheckFloat(dBLoss, 160.05)
		if err != nil {
			t.Error(err)
		}

		dBLoss = FreeSpaceAttenuationDB(433*MHz, 1e+3)
		err = CheckFloat(dBLoss, 85.178)
		if err != nil {
			t.Error(err)
		}

		dBLoss = FreeSpaceAttenuationDB(433*MHz, 1e+6)
		err = CheckFloat(dBLoss, 145.178)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("Can calculate the distance between two lat/lon locations", func(t *testing.T) {
		lat1, lon1 := -36.8485, 174.7633
		lat2, lon2 := -41.2865, 174.7762

		d := CalculateDistance(lat1, lon1, lat2, lon2, R)

		err := CheckFloat(d, 493.4e+3)
		if err != nil {
			t.Error(err)
		}

	})

	t.Run("Can calculate fresnel points", func(t *testing.T) {

		// Magic Numbers from: http://www.wirelessconnections.net/calcs/FresnelZone.asp

		zone, err := FresnelFirstZoneMax(2.4*GHz, 10e+3)
		if err != nil {
			t.Error(err)
		}
		err = CheckFloat(zone, 17.671776)
		if err != nil {
			t.Error(err)
		}

		zone, err = FresnelFirstZoneMax(2.4*GHz, 100e+3)
		if err != nil {
			t.Error(err)
		}
		err = CheckFloat(zone, 55.88)
		if err != nil {
			t.Error(err)
		}

	})

}
