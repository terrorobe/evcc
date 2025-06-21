package core

import (
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/evcc-io/evcc/core/keys"
	"github.com/evcc-io/evcc/server/db"
	"github.com/evcc-io/evcc/server/db/settings"
	"github.com/evcc-io/evcc/util"
	"github.com/stretchr/testify/require"
)

// TestMinimalRestore - Isolate the restoration issue with absolute minimal setup
func TestMinimalRestore(t *testing.T) {
	// Initialize database
	require.NoError(t, db.NewInstance("sqlite", ":memory:"))
	require.NoError(t, settings.Init())

	// Store test data
	testForecast := 123.456
	testPvData := map[string]float64{"pv1": 78.901}
	settings.SetFloat(keys.SolarAccForecast, testForecast)
	require.NoError(t, settings.SetJson(keys.SolarAccYield, testPvData))

	// Create absolutely minimal site - only the fields we know are needed
	logger := util.NewLogger("test")
	// Enable DEBUG level logging to see the debug prints
	logger.DEBUG.Println("Logger created with DEBUG level")
	
	site := &Site{
		log:        logger,
		pvEnergy:   make(map[string]*meterEnergy),
		fcstEnergy: &meterEnergy{clock: clock.New()},
		Meters: MetersConfig{
			PVMetersRef: []string{"pv1"},
		},
	}

	t.Logf("=== BEFORE CALL ===")
	t.Logf("fcstEnergy.Accumulated = %.6f", site.fcstEnergy.Accumulated)
	t.Logf("pvEnergy map size = %d", len(site.pvEnergy))

	// First call the actual restoreSettings method to see debug output
	t.Logf("=== CALLING ACTUAL restoreSettings() WITH DEBUG PRINTS ===")
	err := site.restoreSettings()
	require.NoError(t, err)
	
	t.Logf("=== REPRODUCING EXACT restoreSettings LOGIC ===")
	
	// This is copied from site.go:290-352, with logging added
	err = func() error {
		t.Logf("STEP: Starting battery settings restoration...")
		
		// All the battery settings restoration (lines 290-318)
		if v, err := settings.Float(keys.BufferSoc); err == nil {
			t.Logf("STEP: Found BufferSoc setting: %.6f", v)
			if err := site.SetBufferSoc(v); err != nil {
				t.Logf("STEP: SetBufferSoc failed: %v", err)
				return err
			}
		}
		if v, err := settings.Float(keys.BufferStartSoc); err == nil {
			t.Logf("STEP: Found BufferStartSoc setting: %.6f", v)
			if err := site.SetBufferStartSoc(v); err != nil {
				t.Logf("STEP: SetBufferStartSoc failed: %v", err)
				return err
			}
		}
		if v, err := settings.Float(keys.PrioritySoc); err == nil {
			t.Logf("STEP: Found PrioritySoc setting: %.6f", v)
			if err := site.SetPrioritySoc(v); err != nil {
				t.Logf("STEP: SetPrioritySoc failed: %v", err)
				return err
			}
		}
		if v, err := settings.Bool(keys.BatteryDischargeControl); err == nil {
			t.Logf("STEP: Found BatteryDischargeControl setting: %v", v)
			if err := site.SetBatteryDischargeControl(v); err != nil {
				t.Logf("STEP: SetBatteryDischargeControl failed: %v", err)
				return err
			}
		}
		if v, err := settings.Float(keys.ResidualPower); err == nil {
			t.Logf("STEP: Found ResidualPower setting: %.6f", v)
			if err := site.SetResidualPower(v); err != nil {
				t.Logf("STEP: SetResidualPower failed: %v", err)
				return err
			}
		}
		if v, err := settings.Float(keys.BatteryGridChargeLimit); err == nil {
			t.Logf("STEP: Found BatteryGridChargeLimit setting: %.6f", v)
			site.SetBatteryGridChargeLimit(&v)
		}

		t.Logf("STEP: Battery settings restoration completed successfully")
		t.Logf("STEP: Starting energy restoration logic...")

		// restore accumulated energy (lines 320-349)
		pvEnergy := make(map[string]float64)
		fcstEnergy, err := settings.Float(keys.SolarAccForecast)
		t.Logf("STEP: settings.Float(SolarAccForecast) = %.6f, err = %v", fcstEnergy, err)

		jsonErr := settings.Json(keys.SolarAccYield, &pvEnergy)
		t.Logf("STEP: settings.Json(SolarAccYield) = %+v, err = %v", pvEnergy, jsonErr)
		
		condition := err == nil && jsonErr == nil
		t.Logf("STEP: Condition (err == nil && jsonErr == nil) = %v", condition)

		if condition {
			t.Logf("STEP: ENTERING energy restoration block")
			var nok bool
			for _, name := range site.Meters.PVMetersRef {
				t.Logf("STEP: Processing meter '%s'", name)
				if fcst, ok := pvEnergy[name]; ok {
					t.Logf("STEP: Found data for '%s': %.6f", name, fcst)
					t.Logf("STEP: About to execute: site.pvEnergy[name].Accumulated = fcst")
					
					// This should panic!
					site.pvEnergy[name].Accumulated = fcst
					t.Logf("STEP: *** NO PANIC OCCURRED - THIS IS IMPOSSIBLE! ***")
				} else {
					nok = true
					t.Logf("STEP: No data for '%s', setting nok=true", name)
				}
			}

			if !nok {
				site.fcstEnergy.Accumulated = fcstEnergy
				t.Logf("STEP: Restoration successful, set fcstEnergy.Accumulated = %.6f", fcstEnergy)
			} else {
				t.Logf("STEP: Restoration failed, resetting metrics")
				settings.Delete(keys.SolarAccForecast)
				settings.Delete(keys.SolarAccYield)
				for _, pe := range site.pvEnergy {
					pe.Accumulated = 0
				}
			}
		} else {
			t.Logf("STEP: SKIPPING energy restoration block due to condition")
		}

		t.Logf("STEP: Energy restoration logic completed")
		return nil
	}()
	
	require.NoError(t, err)

	t.Logf("=== AFTER CALL ===")
	t.Logf("fcstEnergy.Accumulated = %.6f (should be %.6f)", site.fcstEnergy.Accumulated, testForecast)
	t.Logf("pvEnergy map size = %d", len(site.pvEnergy))

	// Check if restoration worked
	if site.fcstEnergy.Accumulated == testForecast {
		t.Logf("SUCCESS: Restoration worked!")
	} else if site.fcstEnergy.Accumulated == 0 {
		t.Logf("FAILURE: Restoration didn't work - investigating...")
		
		// Check if data is still there
		storedValue, err := settings.Float(keys.SolarAccForecast)
		t.Logf("Stored value still exists: %.6f (err: %v)", storedValue, err)
		
		// This means the restoration logic was skipped entirely
		t.Logf("CONCLUSION: Energy restoration logic never executed")
		t.Logf("This confirms there's a path through restoreSettings that skips it")
	}
}