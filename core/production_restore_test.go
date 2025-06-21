package core

import (
	"testing"

	"github.com/benbjohnson/clock"
	"github.com/evcc-io/evcc/core/keys"
	"github.com/evcc-io/evcc/server/db"
	"github.com/evcc-io/evcc/server/db/settings"
	"github.com/evcc-io/evcc/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// productionRestoreSettings is a copy of the restoreSettings method but without the testing.Testing() check
func (site *Site) productionRestoreSettings() error {
	site.log.ERROR.Printf("*** PRODUCTION TEST: productionRestoreSettings() method called")
	
	// Skip the testing.Testing() check to run actual restoration logic
	
	if v, err := settings.Float(keys.BufferSoc); err == nil {
		if err := site.SetBufferSoc(v); err != nil {
			return err
		}
	}
	if v, err := settings.Float(keys.BufferStartSoc); err == nil {
		if err := site.SetBufferStartSoc(v); err != nil {
			return err
		}
	}
	if v, err := settings.Float(keys.PrioritySoc); err == nil {
		if err := site.SetPrioritySoc(v); err != nil {
			return err
		}
	}
	if v, err := settings.Bool(keys.BatteryDischargeControl); err == nil {
		if err := site.SetBatteryDischargeControl(v); err != nil {
			return err
		}
	}
	if v, err := settings.Float(keys.ResidualPower); err == nil {
		if err := site.SetResidualPower(v); err != nil {
			return err
		}
	}
	if v, err := settings.Float(keys.BatteryGridChargeLimit); err == nil {
		site.SetBatteryGridChargeLimit(&v)
	}

	// restore accumulated energy - ACTUAL PRODUCTION LOGIC
	site.log.ERROR.Printf("*** PRODUCTION TEST: Starting energy restoration logic")
	site.log.ERROR.Printf("*** PRODUCTION TEST: site.pvEnergy map pointer: %p", site.pvEnergy)
	site.log.ERROR.Printf("*** PRODUCTION TEST: site.pvEnergy map length: %d", len(site.pvEnergy))
	
	pvEnergy := make(map[string]float64)
	fcstEnergy, err := settings.Float(keys.SolarAccForecast)
	site.log.ERROR.Printf("*** PRODUCTION TEST: settings.Float(SolarAccForecast) = %.6f, err = %v", fcstEnergy, err)

	jsonErr := settings.Json(keys.SolarAccYield, &pvEnergy)
	site.log.ERROR.Printf("*** PRODUCTION TEST: settings.Json(SolarAccYield) = %+v, err = %v", pvEnergy, jsonErr)
	
	condition := err == nil && jsonErr == nil
	site.log.ERROR.Printf("*** PRODUCTION TEST: Condition check (err == nil && jsonErr == nil) = %v", condition)

	if condition {
		site.log.ERROR.Printf("*** PRODUCTION TEST: ENTERING restoration block")
		site.log.ERROR.Printf("*** PRODUCTION TEST: site.Meters.PVMetersRef = %v", site.Meters.PVMetersRef)
		site.log.ERROR.Printf("*** PRODUCTION TEST: site.pvEnergy map has %d entries", len(site.pvEnergy))
		
		var nok bool
		for i, name := range site.Meters.PVMetersRef {
			site.log.ERROR.Printf("*** PRODUCTION TEST: Loop iteration %d: processing meter '%s'", i, name)
			
			if fcst, ok := pvEnergy[name]; ok {
				site.log.ERROR.Printf("*** PRODUCTION TEST: Found stored value %.6f for meter '%s'", fcst, name)
				site.log.ERROR.Printf("*** PRODUCTION TEST: Looking up site.pvEnergy['%s']...", name)
				
				entry := site.pvEnergy[name]
				site.log.ERROR.Printf("*** PRODUCTION TEST: site.pvEnergy['%s'] = %p", name, entry)
				
				if entry == nil {
					site.log.ERROR.Printf("*** PRODUCTION TEST: *** site.pvEnergy['%s'] IS NIL - WILL CAUSE PANIC! ***", name)
					site.log.ERROR.Printf("*** PRODUCTION TEST: Map contents dump:")
					for k, v := range site.pvEnergy {
						site.log.ERROR.Printf("*** PRODUCTION TEST:   pvEnergy['%s'] = %p", k, v)
					}
					site.log.ERROR.Printf("*** PRODUCTION TEST: About to access .Accumulated on nil pointer...")
				} else {
					site.log.ERROR.Printf("*** PRODUCTION TEST: site.pvEnergy['%s'] exists at %p, safe to access", name, entry)
				}
				
				site.log.ERROR.Printf("*** PRODUCTION TEST: About to execute: site.pvEnergy['%s'].Accumulated = %.6f", name, fcst)
				site.pvEnergy[name].Accumulated = fcst  // THIS SHOULD PANIC if entry is nil
				site.log.ERROR.Printf("*** PRODUCTION TEST: Successfully set site.pvEnergy['%s'].Accumulated = %.6f", name, fcst)
			} else {
				nok = true
				site.log.ERROR.Printf("*** PRODUCTION TEST: No stored value found for meter '%s', setting nok=true", name)
				site.log.WARN.Printf("accumulated solar yield: cannot restore %s", name)
			}
		}

		site.log.ERROR.Printf("*** PRODUCTION TEST: Loop completed, nok = %v", nok)

		if !nok {
			site.fcstEnergy.Accumulated = fcstEnergy
			site.log.ERROR.Printf("*** PRODUCTION TEST: Restoration SUCCESS - set fcstEnergy.Accumulated = %.6f", fcstEnergy)
			site.log.DEBUG.Printf("accumulated solar yield: restored %.3fkWh forecasted, %+v produced", fcstEnergy, pvEnergy)
		} else {
			// reset metrics
			site.log.ERROR.Printf("*** PRODUCTION TEST: Restoration FAILED - resetting metrics")
			site.log.WARN.Printf("accumulated solar yield: metrics reset")

			settings.Delete(keys.SolarAccForecast)
			settings.Delete(keys.SolarAccYield)

			for _, pe := range site.pvEnergy {
				pe.Accumulated = 0
			}
		}
	} else {
		site.log.ERROR.Printf("*** PRODUCTION TEST: SKIPPING restoration block due to condition failure")
	}
	
	site.log.ERROR.Printf("*** PRODUCTION TEST: Energy restoration logic completed")

	return nil
}

func TestProductionModeRestoration(t *testing.T) {
	// Setup database and settings
	require.NoError(t, db.NewInstance("sqlite", ":memory:"))
	require.NoError(t, settings.Init())
	
	// Store test data like production would
	settings.SetFloat(keys.SolarAccForecast, 365.718)
	require.NoError(t, settings.SetJson(keys.SolarAccYield, map[string]float64{
		"pv1": 271.752,
	}))
	
	// Create site in the EXACT same state as production bug
	site := &Site{
		log:        util.NewLogger("test"),
		pvEnergy:   make(map[string]*meterEnergy), // EMPTY - this is the bug condition
		fcstEnergy: &meterEnergy{clock: clock.New()},
		Meters: MetersConfig{
			PVMetersRef: []string{"pv1"}, // References pv1 but pvEnergy["pv1"] doesn't exist
		},
	}
	
	t.Logf("=== PRODUCTION TEST: Testing actual restoration failure ===")
	t.Logf("Before restoration: pvEnergy map has %d entries", len(site.pvEnergy))
	t.Logf("Before restoration: fcstEnergy.Accumulated = %.6f", site.fcstEnergy.Accumulated)
	
	// This should demonstrate the exact production failure
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Logf("*** EXPECTED PANIC CAUGHT: %v", r)
				t.Logf("This is exactly what happens in production!")
			}
		}()
		
		err := site.productionRestoreSettings()
		if err != nil {
			t.Logf("Error from productionRestoreSettings: %v", err)
		}
	}()
	
	t.Logf("After restoration attempt: fcstEnergy.Accumulated = %.6f", site.fcstEnergy.Accumulated)
	t.Logf("After restoration attempt: pvEnergy map has %d entries", len(site.pvEnergy))
	
	// Now test with proper initialization (like Boot() would do)
	t.Logf("=== PRODUCTION TEST: Testing with proper initialization ===")
	
	// Initialize pvEnergy map like Boot() does
	for _, ref := range site.Meters.PVMetersRef {
		site.pvEnergy[ref] = &meterEnergy{clock: clock.New()}
		t.Logf("Created pvEnergy['%s'] = %p", ref, site.pvEnergy[ref])
	}
	
	// Reset data
	settings.SetFloat(keys.SolarAccForecast, 365.718)
	require.NoError(t, settings.SetJson(keys.SolarAccYield, map[string]float64{
		"pv1": 271.752,
	}))
	
	err := site.productionRestoreSettings()
	require.NoError(t, err)
	
	t.Logf("With proper init: fcstEnergy.Accumulated = %.6f", site.fcstEnergy.Accumulated)
	assert.Equal(t, 365.718, site.fcstEnergy.Accumulated, "fcstEnergy should be restored")
	assert.Equal(t, 271.752, site.pvEnergy["pv1"].Accumulated, "pvEnergy should be restored")
}