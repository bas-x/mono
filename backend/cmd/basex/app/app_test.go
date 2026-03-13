package app

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bas-x/basex/services"
)

func newTestRuntime(t *testing.T) *Runtime {
	t.Helper()

	runtime, err := New(RuntimeConfig{
		Config: Config{
			WindowWidth:  1280,
			WindowHeight: 800,
			WindowTitle:  "test",
		},
		Service: services.NewSimulationService(services.SimulationServiceConfig{}),
	})
	require.NoError(t, err)
	t.Cleanup(runtime.Close)
	return runtime
}

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
		Threat:       services.Threat{ID: "th1", Position: services.Point{X: 10, Y: 20}},
		Timestamp:    time.Unix(0, 1),
	}))
	require.False(t, eventRequiresUIRefresh(services.AllAircraftPositionsEvent{
		Type:         services.EventTypeAllAircraftPositions,
		SimulationID: services.BaseSimulationID,
		Tick:         1,
		Timestamp:    time.Unix(0, 1),
	}))
}

func TestBranchStartupAutoCreatesBaseSimulation(t *testing.T) {
	t.Parallel()

	runtime := newTestRuntime(t)

	require.NotNil(t, runtime.state.Simulation)
	require.Equal(t, services.BaseSimulationID, runtime.state.Simulation.ID)
	require.NotEqual(t, "idle", runtime.state.Status)
}

func TestBranchInitialActiveTabIsBase(t *testing.T) {
	t.Parallel()

	runtime := newTestRuntime(t)

	require.Equal(t, services.BaseSimulationID, runtime.activeSimulationID)
	require.Len(t, runtime.tabs, 1)
	require.Equal(t, "Base", runtime.tabs[0].Label)
	require.Equal(t, services.BaseSimulationID, runtime.tabs[0].SimulationID)
}

func TestBranchCreateFromBaseAddsFriendlyTab(t *testing.T) {
	t.Parallel()

	runtime := newTestRuntime(t)

	branchID, err := runtime.createBranchFromActiveTab()
	require.NoError(t, err)
	require.NotEmpty(t, branchID)
	require.Len(t, runtime.tabs, 2)
	require.Equal(t, "Branch 1", runtime.tabs[1].Label)
	require.Equal(t, branchID, runtime.tabs[1].SimulationID)
}

func TestBranchNonBaseTabsCannotCreateMoreBranches(t *testing.T) {
	t.Parallel()

	runtime := newTestRuntime(t)

	branchID, err := runtime.createBranchFromActiveTab()
	require.NoError(t, err)
	runtime.activeSimulationID = branchID

	_, err = runtime.createBranchFromActiveTab()
	require.Error(t, err)
}

func TestBranchResetRemovesBranchesAndRecreatesBase(t *testing.T) {
	t.Parallel()

	runtime := newTestRuntime(t)

	_, err := runtime.createBranchFromActiveTab()
	require.NoError(t, err)
	require.Len(t, runtime.tabs, 2)

	require.NoError(t, runtime.resetSimulation())

	require.Len(t, runtime.tabs, 1)
	require.Equal(t, "Base", runtime.tabs[0].Label)
	require.Equal(t, services.BaseSimulationID, runtime.tabs[0].SimulationID)
	require.Equal(t, services.BaseSimulationID, runtime.activeSimulationID)
	require.NotNil(t, runtime.state.Simulation)
	require.Equal(t, services.BaseSimulationID, runtime.state.Simulation.ID)
}
