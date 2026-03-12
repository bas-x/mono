package app

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bas-x/basex/services"
)

func TestEventRequiresUIRefresh(t *testing.T) {
	t.Parallel()

	require.False(t, eventRequiresUIRefresh(nil))
	require.False(t, eventRequiresUIRefresh(services.SimulationStepEvent{
		Type:         services.EventTypeSimulationStep,
		SimulationID: services.BaseSimulationID,
		Tick:         12,
		Timestamp:    time.Unix(0, 1),
	}))
	require.True(t, eventRequiresUIRefresh(services.AircraftStateChangeEvent{
		Type:         services.EventTypeAircraftStateChange,
		SimulationID: services.BaseSimulationID,
		TailNumber:   "abcd",
		OldState:     "Outbound",
		NewState:     "Inbound",
		Timestamp:    time.Unix(0, 1),
	}))
	require.False(t, eventRequiresUIRefresh(services.LandingAssignmentEvent{
		Type:         services.EventTypeLandingAssignment,
		SimulationID: services.BaseSimulationID,
		TailNumber:   "abcd",
		BaseID:       "efgh",
		Timestamp:    time.Unix(0, 1),
	}))
	require.False(t, eventRequiresUIRefresh(services.ThreatSpawnedEvent{
		Type:         services.EventTypeThreatSpawned,
		SimulationID: services.BaseSimulationID,
		Threat:       services.Threat{ID: "th1", RegionID: "SE-K", Region: "Blekinge"},
		Timestamp:    time.Unix(0, 1),
	}))
}
