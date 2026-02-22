package api

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanStrategyUnmarshalDefaults(t *testing.T) {
	var s PlanStrategy
	require.NoError(t, json.Unmarshal([]byte(`{"continuous":true,"precondition":900}`), &s))

	assert.True(t, s.Continuous)
	assert.Equal(t, 15*time.Minute, s.Precondition)
	assert.Equal(t, PlanPowerMax, s.Power)
}

func TestPlanStrategyUnmarshalPower(t *testing.T) {
	var s PlanStrategy
	require.NoError(t, json.Unmarshal([]byte(`{"continuous":false,"precondition":0,"power":"required"}`), &s))

	assert.Equal(t, PlanPowerRequired, s.Power)
}

func TestPlanStrategyUnmarshalInvalidPower(t *testing.T) {
	var s PlanStrategy
	require.NoError(t, json.Unmarshal([]byte(`{"continuous":false,"precondition":0,"power":"bogus"}`), &s))

	assert.Equal(t, PlanPowerMax, s.Power)
}

func TestPlanStrategyMarshalDefaultPower(t *testing.T) {
	data, err := json.Marshal(PlanStrategy{Continuous: false, Precondition: time.Minute})
	require.NoError(t, err)

	assert.Contains(t, string(data), `"power":"max"`)
}
