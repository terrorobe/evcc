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

// TestSimpleDebug - Just call restoreSettings and see what debug output we get
func TestSimpleDebug(t *testing.T) {
	// Initialize database
	require.NoError(t, db.NewInstance("sqlite", ":memory:"))
	require.NoError(t, settings.Init())

	// Store test data exactly like the user's scenario
	settings.SetFloat(keys.SolarAccForecast, 365.718)
	require.NoError(t, settings.SetJson(keys.SolarAccYield, map[string]float64{"pv1": 271.752}))

	// Create minimal site
	site := &Site{
		log:        util.NewLogger("test"),
		pvEnergy:   make(map[string]*meterEnergy), // Empty map - this is the bug condition
		fcstEnergy: &meterEnergy{clock: clock.New()},
		Meters: MetersConfig{
			PVMetersRef: []string{"pv1"}, // References pv1 but pvEnergy["pv1"] doesn't exist
		},
	}

	t.Logf("=== CALLING restoreSettings() ===")
	t.Logf("Before: fcstEnergy.Accumulated = %.6f", site.fcstEnergy.Accumulated)
	t.Logf("Before: pvEnergy map size = %d", len(site.pvEnergy))
	
	// This should trigger our debug prints
	err := site.restoreSettings()
	require.NoError(t, err)
	
	t.Logf("After: fcstEnergy.Accumulated = %.6f", site.fcstEnergy.Accumulated)
	t.Logf("After: pvEnergy map size = %d", len(site.pvEnergy))
}