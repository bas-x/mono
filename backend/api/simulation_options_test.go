package api

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMapSimulationOptionsRequest_MergesMissingGroupsWithDefaults(t *testing.T) {
	t.Parallel()

	defaults := defaultBaseSimulationOptions()
	req := &simulationOptionsRequest{
		ConstellationOpts: &constellationOptionsRequest{
			IncludeRegions: []string{"Blekinge"},
			MinPerRegion:   1,
			MaxPerRegion:   1,
			MaxTotal:       1,
		},
		FleetOpts: &fleetOptionsRequest{
			AircraftMin: 1,
			AircraftMax: 1,
			NeedsMin:    0,
			NeedsMax:    0,
		},
	}

	options, err := mapSimulationOptionsRequest(req)
	require.NoError(t, err)
	require.NotNil(t, options)

	require.Equal(t, defaults.ThreatOpts.MaxActive, options.ThreatOpts.MaxActive)
	require.Equal(t, defaults.ThreatOpts.MaxActiveTicks, options.ThreatOpts.MaxActiveTicks)
	require.Equal(t, defaults.ThreatOpts.SpawnChance.Numerator(), options.ThreatOpts.SpawnChance.Numerator())
	require.Equal(t, defaults.ThreatOpts.SpawnChance.Denominator(), options.ThreatOpts.SpawnChance.Denominator())
}

func TestMapSimulationOptionsRequest_OverridesThreatOptionsWhenProvided(t *testing.T) {
	t.Parallel()

	req := &simulationOptionsRequest{
		ThreatOpts: &threatOptionsRequest{
			SpawnChance: &ratioRequest{Numerator: 1, Denominator: 1},
			MaxActive:   7,
		},
	}

	options, err := mapSimulationOptionsRequest(req)
	require.NoError(t, err)
	require.NotNil(t, options)

	require.Equal(t, uint(7), options.ThreatOpts.MaxActive)
	require.Equal(t, uint64(0), options.ThreatOpts.MaxActiveTicks)
	require.Equal(t, uint64(1), options.ThreatOpts.SpawnChance.Numerator())
	require.Equal(t, uint64(1), options.ThreatOpts.SpawnChance.Denominator())
}
