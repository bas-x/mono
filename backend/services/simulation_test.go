package services

import (
	"encoding/hex"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bas-x/basex/prng"
	"github.com/bas-x/basex/simulation"
)

func TestSimulationService_AirbasesAndAircrafts(t *testing.T) {
	t.Parallel()

	svc := NewSimulationService(SimulationServiceConfig{})
	_, err := svc.CreateBaseSimulation(BaseSimulationConfig{
		Options: &simulation.SimulationOptions{
			ConstellationOpts: simulation.ConstellationOptions{
				IncludeRegions:    []string{"Blekinge"},
				MinPerRegion:      1,
				MaxPerRegion:      1,
				MaxTotal:          1,
				RegionProbability: prng.New(1, 1),
			},
			FleetOpts: simulation.FleetOptions{
				AircraftMin: 1,
				AircraftMax: 1,
			},
		},
	})
	require.NoError(t, err)

	bases, err := svc.Airbases(BaseSimulationID)
	require.NoError(t, err)
	require.Len(t, bases, 1)
	require.Len(t, bases[0].ID, 16)
	_, err = hex.DecodeString(bases[0].ID)
	require.NoError(t, err)

	aircrafts, err := svc.Aircrafts(BaseSimulationID)
	require.NoError(t, err)
	require.Len(t, aircrafts, 1)
	require.Len(t, aircrafts[0].TailNumber, 16)
	_, err = hex.DecodeString(aircrafts[0].TailNumber)
	require.NoError(t, err)
	require.NotEmpty(t, aircrafts[0].State)
}

func TestSimulationService_UnknownSimulationID(t *testing.T) {
	t.Parallel()

	svc := NewSimulationService(SimulationServiceConfig{})

	_, err := svc.Airbases("branch-1")
	require.ErrorIs(t, err, ErrSimulationNotFound)

	_, err = svc.Aircrafts("branch-1")
	require.ErrorIs(t, err, ErrSimulationNotFound)
}

func TestSimulationService_SimulationsEmptyBeforeCreate(t *testing.T) {
	t.Parallel()

	svc := NewSimulationService(SimulationServiceConfig{})
	require.Empty(t, svc.Simulations())

	base, ok := svc.Base()
	require.False(t, ok)
	require.Nil(t, base)
}

func TestSimulationService_StartSimulationErrors(t *testing.T) {
	t.Parallel()

	svc := NewSimulationService(SimulationServiceConfig{})
	require.ErrorIs(t, svc.StartSimulation(BaseSimulationID), ErrBaseNotFound)
	require.ErrorIs(t, svc.StartSimulation("branch-1"), ErrSimulationNotFound)

	_, err := svc.CreateBaseSimulation(BaseSimulationConfig{Options: testSimulationOptions(1, 1)})
	require.NoError(t, err)

	require.ErrorIs(t, svc.StartSimulation("branch-1"), ErrSimulationNotFound)
}

func TestSimulationService_SimulationsReflectRunningState(t *testing.T) {
	t.Parallel()

	svc := NewSimulationService(SimulationServiceConfig{RunnerConfig: simulation.ControlledRunnerConfig{TicksPerSecond: 128}})
	_, err := svc.CreateBaseSimulation(BaseSimulationConfig{Options: testSimulationOptions(1, 1)})
	require.NoError(t, err)

	list := svc.Simulations()
	require.Len(t, list, 1)
	require.Equal(t, BaseSimulationID, list[0].ID)
	require.False(t, list[0].Running)

	require.NoError(t, svc.StartSimulation(BaseSimulationID))
	require.Eventually(t, func() bool {
		listed := svc.Simulations()
		return len(listed) == 1 && listed[0].Running
	}, time.Second, 10*time.Millisecond)
}

func TestSimulationService_StartSimulationStartsAllBranches(t *testing.T) {
	t.Parallel()

	svc := NewSimulationService(SimulationServiceConfig{RunnerConfig: simulation.ControlledRunnerConfig{TicksPerSecond: 128}})
	_, err := svc.CreateBaseSimulation(BaseSimulationConfig{Options: testSimulationOptions(1, 1)})
	require.NoError(t, err)

	branchID, err := svc.BranchSimulation(BaseSimulationID)
	require.NoError(t, err)

	_, eventCh := svc.Broadcaster().Subscribe()
	require.NoError(t, svc.StartSimulation(branchID))

	require.Eventually(t, func() bool {
		baseInfo, err := svc.Simulation(BaseSimulationID)
		if err != nil {
			return false
		}
		branchInfo, err := svc.Simulation(branchID)
		if err != nil {
			return false
		}
		return baseInfo.Running && !baseInfo.Paused && branchInfo.Running && !branchInfo.Paused
	}, time.Second, 10*time.Millisecond)

	baseStep := waitForSimulationStepEvent(t, eventCh, BaseSimulationID, time.Second)
	branchStep := waitForSimulationStepEvent(t, eventCh, branchID, time.Second)
	require.Greater(t, baseStep.Tick, uint64(0))
	require.Greater(t, branchStep.Tick, uint64(0))

	require.ErrorIs(t, svc.StartSimulation(BaseSimulationID), ErrSimulationRunning)
}

func TestSimulationService_PauseResume(t *testing.T) {
	t.Parallel()

	svc := NewSimulationService(SimulationServiceConfig{RunnerConfig: simulation.ControlledRunnerConfig{TicksPerSecond: 128}})
	_, err := svc.CreateBaseSimulation(BaseSimulationConfig{Options: testSimulationOptions(1, 1)})
	require.NoError(t, err)

	_, eventCh := svc.Broadcaster().Subscribe()
	require.NoError(t, svc.StartSimulation(BaseSimulationID))

	firstTick := waitForSimulationStepEvent(t, eventCh, BaseSimulationID, time.Second)
	require.Greater(t, firstTick.Tick, uint64(0))

	branchID, err := svc.BranchSimulation(BaseSimulationID)
	require.NoError(t, err)

	baseInfo, err := svc.Simulation(BaseSimulationID)
	require.NoError(t, err)
	branchInfo, err := svc.Simulation(branchID)
	require.NoError(t, err)
	require.True(t, baseInfo.Running)
	require.True(t, baseInfo.Paused)
	require.True(t, branchInfo.Running)
	require.True(t, branchInfo.Paused)

	require.NoError(t, svc.ResumeSimulation(branchID))

	baseInfo, err = svc.Simulation(BaseSimulationID)
	require.NoError(t, err)
	branchInfo, err = svc.Simulation(branchID)
	require.NoError(t, err)
	require.True(t, baseInfo.Running)
	require.False(t, baseInfo.Paused)
	require.True(t, branchInfo.Running)
	require.False(t, branchInfo.Paused)

	nextBaseTick := waitForSimulationStepEvent(t, eventCh, BaseSimulationID, time.Second)
	require.Greater(t, nextBaseTick.Tick, firstTick.Tick)
	nextBranchTick := waitForSimulationStepEvent(t, eventCh, branchID, time.Second)
	require.Greater(t, nextBranchTick.Tick, uint64(0))

	require.NoError(t, svc.PauseSimulation(branchID))
	baseInfo, err = svc.Simulation(BaseSimulationID)
	require.NoError(t, err)
	branchInfo, err = svc.Simulation(branchID)
	require.NoError(t, err)
	require.True(t, baseInfo.Running)
	require.True(t, baseInfo.Paused)
	require.True(t, branchInfo.Running)
	require.True(t, branchInfo.Paused)

	drainStepEvents(eventCh)
	ensureNoStepEvent(t, eventCh, 150*time.Millisecond)
}

func TestSimulationService_Simulation(t *testing.T) {
	t.Parallel()

	svc := NewSimulationService(SimulationServiceConfig{})
	_, err := svc.Simulation(BaseSimulationID)
	require.ErrorIs(t, err, ErrBaseNotFound)

	_, err = svc.CreateBaseSimulation(BaseSimulationConfig{Options: testSimulationOptions(1, 1)})
	require.NoError(t, err)

	info, err := svc.Simulation(BaseSimulationID)
	require.NoError(t, err)
	require.Equal(t, BaseSimulationID, info.ID)
	require.False(t, info.Running)

	_, err = svc.Simulation("branch-1")
	require.ErrorIs(t, err, ErrSimulationNotFound)
}

func TestSimulationService_Threats(t *testing.T) {
	t.Parallel()

	svc := NewSimulationService(SimulationServiceConfig{})
	_, err := svc.CreateBaseSimulation(BaseSimulationConfig{Options: &simulation.SimulationOptions{
		ConstellationOpts: simulation.ConstellationOptions{
			IncludeRegions:    []string{"Blekinge"},
			MinPerRegion:      1,
			MaxPerRegion:      1,
			MaxTotal:          1,
			RegionProbability: prng.New(1, 1),
		},
		ThreatOpts: simulation.ThreatOptions{SpawnChance: prng.New(1, 1), MaxActive: 2},
	}})
	require.NoError(t, err)

	base, ok := svc.Base()
	require.True(t, ok)
	base.Step()

	threats, err := svc.Threats(BaseSimulationID)
	require.NoError(t, err)
	require.NotEmpty(t, threats)
	require.NotZero(t, threats[0].Position.X+threats[0].Position.Y)
}

func TestSimulationService_ResetClearsSimulationAndStopsRunner(t *testing.T) {
	t.Parallel()

	svc := NewSimulationService(SimulationServiceConfig{RunnerConfig: simulation.ControlledRunnerConfig{TicksPerSecond: 128}})
	_, err := svc.CreateBaseSimulation(BaseSimulationConfig{Options: testSimulationOptions(1, 1)})
	require.NoError(t, err)
	require.NoError(t, svc.StartSimulation(BaseSimulationID))

	svc.Reset()

	require.Empty(t, svc.Simulations())
	base, ok := svc.Base()
	require.False(t, ok)
	require.Nil(t, base)
	require.ErrorIs(t, svc.StartSimulation(BaseSimulationID), ErrBaseNotFound)
	require.ErrorIs(t, svc.StartSimulation("branch-1"), ErrSimulationNotFound)
}

func TestSimulationService_ResetSimulationErrors(t *testing.T) {
	t.Parallel()

	svc := NewSimulationService(SimulationServiceConfig{})
	require.ErrorIs(t, svc.ResetSimulation(BaseSimulationID), ErrBaseNotFound)
	require.ErrorIs(t, svc.ResetSimulation("branch-1"), ErrSimulationNotFound)

	_, err := svc.CreateBaseSimulation(BaseSimulationConfig{Options: testSimulationOptions(1, 1)})
	require.NoError(t, err)

	require.ErrorIs(t, svc.ResetSimulation("branch-1"), ErrSimulationNotFound)
	require.NoError(t, svc.ResetSimulation(BaseSimulationID))
	require.Empty(t, svc.Simulations())
}

func TestSimulationService_RunnerStopsWhenUntilTickReached(t *testing.T) {
	t.Parallel()

	svc := NewSimulationService(SimulationServiceConfig{
		RunnerConfig: simulation.ControlledRunnerConfig{TicksPerSecond: 512},
		RunUntilTick: 3,
	})
	_, err := svc.CreateBaseSimulation(BaseSimulationConfig{Options: testSimulationOptions(1, 1)})
	require.NoError(t, err)

	require.NoError(t, svc.StartSimulation(BaseSimulationID))
	require.Eventually(t, func() bool {
		listed := svc.Simulations()
		return len(listed) == 1 && !listed[0].Running
	}, 2*time.Second, 10*time.Millisecond)
}

func testSimulationOptions(numBases, numAircraft uint) *simulation.SimulationOptions {
	return &simulation.SimulationOptions{
		ConstellationOpts: simulation.ConstellationOptions{
			IncludeRegions:    []string{"Blekinge"},
			MinPerRegion:      numBases,
			MaxPerRegion:      numBases,
			MaxTotal:          numBases,
			RegionProbability: prng.New(1, 1),
		},
		FleetOpts: simulation.FleetOptions{
			AircraftMin: numAircraft,
			AircraftMax: numAircraft,
			NeedsMin:    1,
			NeedsMax:    2,
		},
	}
}

func waitForStepEvent(t *testing.T, ch <-chan Event, timeout time.Duration) SimulationStepEvent {
	t.Helper()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case event := <-ch:
			if step, ok := event.(SimulationStepEvent); ok {
				return step
			}
		case <-timer.C:
			t.Fatal("timed out waiting for simulation step event")
		}
	}
}

func waitForSimulationStepEvent(t *testing.T, ch <-chan Event, simulationID string, timeout time.Duration) SimulationStepEvent {
	t.Helper()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case event := <-ch:
			step, ok := event.(SimulationStepEvent)
			if !ok || step.SimulationID != simulationID {
				continue
			}
			return step
		case <-timer.C:
			t.Fatalf("timed out waiting for simulation step event for %s", simulationID)
		}
	}
}

func ensureNoStepEvent(t *testing.T, ch <-chan Event, duration time.Duration) {
	t.Helper()
	timer := time.NewTimer(duration)
	defer timer.Stop()
	for {
		select {
		case event := <-ch:
			if _, ok := event.(SimulationStepEvent); ok {
				t.Fatalf("unexpected simulation step event while paused: %#v", event)
			}
		case <-timer.C:
			return
		}
	}
}

func drainStepEvents(ch <-chan Event) {
	for {
		select {
		case <-ch:
			continue
		default:
			return
		}
	}
}
