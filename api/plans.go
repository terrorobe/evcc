package api

import (
	"encoding/json"
	"time"
)

type RepeatingPlan struct {
	Weekdays []int  `json:"weekdays"` // 0-6 (Sunday-Saturday)
	Time     string `json:"time"`     // HH:MM
	Tz       string `json:"tz"`       // timezone in IANA format
	Soc      int    `json:"soc"`      // target soc
	Active   bool   `json:"active"`   // active flag
}

type PlanPowerMode string

const (
	PlanPowerMax      PlanPowerMode = "max"      // charge with effective maximum current in active plan slots
	PlanPowerRequired PlanPowerMode = "required" // modulate current to follow required average plan power
)

type PlanStrategy struct {
	Continuous   bool          `json:"continuous"`   // force continuous planning
	Precondition time.Duration `json:"precondition"` // precondition duration in seconds
	Power        PlanPowerMode `json:"power"`        // plan execution power strategy
}

type planStrategy struct {
	Continuous   bool          `json:"continuous"`   // force continuous planning
	Precondition int64         `json:"precondition"` // precondition duration in seconds
	Power        PlanPowerMode `json:"power"`
}

func (ps PlanStrategy) normalized() PlanStrategy {
	res := ps
	if res.Power != PlanPowerRequired {
		res.Power = PlanPowerMax
	}
	if res.Precondition < 0 {
		res.Precondition = 0
	}
	return res
}

func (ps PlanStrategy) MarshalJSON() ([]byte, error) {
	ps = ps.normalized()

	return json.Marshal(planStrategy{
		Continuous:   ps.Continuous,
		Precondition: int64(ps.Precondition.Seconds()),
		Power:        ps.Power,
	})
}

func (ps *PlanStrategy) UnmarshalJSON(data []byte) error {
	var res planStrategy
	if err := json.Unmarshal(data, &res); err != nil {
		return err
	}

	*ps = PlanStrategy{
		Continuous:   res.Continuous,
		Precondition: time.Duration(res.Precondition) * time.Second,
		Power:        res.Power,
	}.normalized()

	return nil
}
