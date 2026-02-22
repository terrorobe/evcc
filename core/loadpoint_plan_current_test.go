package core

import (
	"testing"
	"time"

	evbus "github.com/asaskevich/EventBus"
	"github.com/benbjohnson/clock"
	"github.com/evcc-io/evcc/api"
	"github.com/evcc-io/evcc/core/circuit"
	"github.com/evcc-io/evcc/util"
	"go.uber.org/mock/gomock"
)

func setTestVoltage(t *testing.T) {
	t.Helper()

	oldVoltage := Voltage
	Voltage = 230
	t.Cleanup(func() {
		Voltage = oldVoltage
	})
}

func TestPlanChargingRequiredCurrentSwitchesToThreePhases(t *testing.T) {
	setTestVoltage(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	phaseSwitcher := api.NewMockPhaseSwitcher(ctrl)
	phaseSwitcher.EXPECT().Phases1p3p(3).Return(nil)

	charger := struct {
		*api.MockCharger
		*api.MockPhaseSwitcher
	}{
		MockCharger:       api.NewMockCharger(ctrl),
		MockPhaseSwitcher: phaseSwitcher,
	}
	charger.MockCharger.EXPECT().MaxCurrent(int64(7)).Return(nil)

	clk := clock.NewMock()
	lp := &Loadpoint{
		log:              util.NewLogger("foo"),
		bus:              evbus.New(),
		clock:            clk,
		charger:          charger,
		minCurrent:       6,
		maxCurrent:       16,
		phases:           1,
		enabled:          true,
		wakeUpTimer:      NewTimer(),
		planStrategy:     api.PlanStrategy{Power: api.PlanPowerRequired},
		planTime:         clk.Now().Add(4 * time.Hour),
		planEnergy:       20, // kWh
		planEnergyOffset: 0,
	}

	if err := lp.planCharging(); err != nil {
		t.Fatalf("planCharging returned error: %v", err)
	}
}

func TestPlanChargingRequiredCurrentSwitchesToOnePhaseForCircuitLimit(t *testing.T) {
	setTestVoltage(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	phaseSwitcher := api.NewMockPhaseSwitcher(ctrl)
	phaseSwitcher.EXPECT().Phases1p3p(1).Return(nil)

	charger := struct {
		*api.MockCharger
		*api.MockPhaseSwitcher
	}{
		MockCharger:       api.NewMockCharger(ctrl),
		MockPhaseSwitcher: phaseSwitcher,
	}
	charger.MockCharger.EXPECT().MaxCurrent(int64(6)).Return(nil)

	circuit, err := circuit.New(util.NewLogger("foo"), "test", 0, 3000, nil, time.Minute)
	if err != nil {
		t.Fatalf("create circuit: %v", err)
	}

	clk := clock.NewMock()
	lp := &Loadpoint{
		log:              util.NewLogger("foo"),
		bus:              evbus.New(),
		clock:            clk,
		charger:          charger,
		circuit:          circuit,
		minCurrent:       6,
		maxCurrent:       16,
		phases:           3,
		enabled:          true,
		wakeUpTimer:      NewTimer(),
		planStrategy:     api.PlanStrategy{Power: api.PlanPowerRequired},
		planTime:         clk.Now().Add(4 * time.Hour),
		planEnergy:       5, // kWh
		planEnergyOffset: 0,
	}

	if err := lp.planCharging(); err != nil {
		t.Fatalf("planCharging returned error: %v", err)
	}
}

func TestPlanChargingRequiredCurrentClampsToMinAndMax(t *testing.T) {
	setTestVoltage(t)

	testcases := []struct {
		name     string
		energy   float64
		targetIn time.Duration
		expectA  int64
	}{
		{name: "min current", energy: 0.1, targetIn: 4 * time.Hour, expectA: 6},
		{name: "max current", energy: 50, targetIn: 4 * time.Hour, expectA: 16},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			charger := api.NewMockCharger(ctrl)
			charger.EXPECT().MaxCurrent(tc.expectA).Return(nil)

			clk := clock.NewMock()
			lp := &Loadpoint{
				log:              util.NewLogger("foo"),
				bus:              evbus.New(),
				clock:            clk,
				charger:          charger,
				minCurrent:       6,
				maxCurrent:       16,
				phases:           3,
				enabled:          true,
				planStrategy:     api.PlanStrategy{Power: api.PlanPowerRequired},
				planTime:         clk.Now().Add(tc.targetIn),
				planEnergy:       tc.energy,
				planEnergyOffset: 0,
			}

			if err := lp.planCharging(); err != nil {
				t.Fatalf("planCharging returned error: %v", err)
			}
		})
	}
}
