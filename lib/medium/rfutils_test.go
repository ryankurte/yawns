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
		// TODO: these values need to be checked.
		// Can't see why they don't agree, but, different from all the online calculators :-/

		// Tests with fake frequency

		fakeFreq := WavelengthToFrequency(math.Pi * 4)
		fmt.Printf("Fake frequency: %.2f Hz", fakeFreq)

		dBLoss := FreeSpaceAttenuationDB(fakeFreq, 1e+1)
		err := CheckFloat(dBLoss, 20.0)
		if err != nil {
			t.Error(err)
		}

		dBLoss = FreeSpaceAttenuationDB(fakeFreq, 1e+3)
		err = CheckFloat(dBLoss, 60.0)
		if err != nil {
			t.Error(err)
		}

		// Tests with precalculated results

		dBLoss = FreeSpaceAttenuationDB(2.4e+6, 1e+3)
		err = CheckFloat(dBLoss, 40.02)
		if err != nil {
			t.Error(err)
		}

		dBLoss = FreeSpaceAttenuationDB(2.4e+6, 1e+6)
		err = CheckFloat(dBLoss, 100.05)
		if err != nil {
			t.Error(err)
		}

		dBLoss = FreeSpaceAttenuationDB(433e3, 1e+3)
		err = CheckFloat(dBLoss, 25.177541)
		if err != nil {
			t.Error(err)
		}

		dBLoss = FreeSpaceAttenuationDB(433e3, 1e+6)
		err = CheckFloat(dBLoss, 85.177541)
		if err != nil {
			t.Error(err)
		}
	})

	t.Run("Can calculate fresnel points", func(t *testing.T) {

		zone, err := FresnelFirstZoneMax(2.4e+6, 10e+3)
		if err != nil {
			t.Error(err)
		}

		err = CheckFloat(zone, 17.671776)
		if err != nil {
			t.Error(err)
		}

	})

}
