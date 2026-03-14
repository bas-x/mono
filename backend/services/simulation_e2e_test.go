package services_test

import (
	"encoding/hex"
	"math/rand/v2"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bas-x/basex/prng"
	"github.com/bas-x/basex/services"
	"github.com/bas-x/basex/simulation"
)

func TestSimulationServiceEndToEnd_BaseReadModels(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{Resolution: 10 * time.Minute})
	_, err := svc.CreateBaseSimulation(services.BaseSimulationConfig{Options: safeSimulationOptions(2, 2)})
	require.NoError(t, err)

	airbases, err := svc.Airbases(services.BaseSimulationID)
	require.NoError(t, err)
	require.Len(t, airbases, 2)
	for _, airbase := range airbases {
		require.NotEmpty(t, airbase.RegionID)
		require.NotEmpty(t, airbase.Region)
		require.Len(t, airbase.ID, 16)
		_, decodeErr := hex.DecodeString(airbase.ID)
		require.NoError(t, decodeErr)
	}

	aircrafts, err := svc.Aircrafts(services.BaseSimulationID)
	require.NoError(t, err)
	require.Len(t, aircrafts, 2)
	for _, aircraft := range aircrafts {
		require.Len(t, aircraft.TailNumber, 16)
		_, decodeErr := hex.DecodeString(aircraft.TailNumber)
		require.NoError(t, decodeErr)
		require.NotEmpty(t, aircraft.State)
		require.NotNil(t, aircraft.Needs)
		require.NotEmpty(t, aircraft.Needs)
	}
}

func TestSimulationServiceEndToEnd_LifecycleAndSimulationIDHandling(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{Resolution: 10 * time.Minute})

	_, err := svc.Airbases(services.BaseSimulationID)
	require.ErrorIs(t, err, services.ErrBaseNotFound)

	_, err = svc.Aircrafts("branch-a")
	require.ErrorIs(t, err, services.ErrSimulationNotFound)

	_, err = svc.CreateBaseSimulation(services.BaseSimulationConfig{Options: safeSimulationOptions(1, 1)})
	require.NoError(t, err)

	_, err = svc.CreateBaseSimulation(services.BaseSimulationConfig{Options: safeSimulationOptions(1, 1)})
	require.ErrorIs(t, err, services.ErrBaseAlreadyExists)

	_, err = svc.Airbases("branch-a")
	require.ErrorIs(t, err, services.ErrSimulationNotFound)

	svc.Reset()

	_, err = svc.Aircrafts(services.BaseSimulationID)
	require.ErrorIs(t, err, services.ErrBaseNotFound)
}

func TestSimulationServiceEndToEnd_EmitsEvents(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{})
	_, events := svc.Broadcaster().Subscribe()
	_, err := svc.CreateBaseSimulation(services.BaseSimulationConfig{Options: safeSimulationOptions(1, 1)})
	require.NoError(t, err)

	base, ok := svc.Base()
	require.True(t, ok)

	var sawStep bool

	for range 24 {
		base.Step()
	drainLoop:
		for {
			select {
			case raw := <-events:
				switch event := raw.(type) {
				case services.SimulationStepEvent:
					sawStep = true
					require.Equal(t, services.BaseSimulationID, event.SimulationID)
				}
			case <-time.After(25 * time.Millisecond):
				break drainLoop
			}
		}
		if sawStep {
			break
		}
	}

	require.True(t, sawStep)
}

func TestSimulationServiceEndToEnd_StartSimulationAndStatus(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{
		RunnerConfig: simulation.ControlledRunnerConfig{TicksPerSecond: 128},
	})
	_, err := svc.CreateBaseSimulation(services.BaseSimulationConfig{Options: safeSimulationOptions(1, 1)})
	require.NoError(t, err)

	list := svc.Simulations()
	require.Len(t, list, 1)
	require.Equal(t, services.BaseSimulationID, list[0].ID)
	require.False(t, list[0].Running)

	_, events := svc.Broadcaster().Subscribe()
	require.NoError(t, svc.StartSimulation(services.BaseSimulationID))
	require.ErrorIs(t, svc.StartSimulation(services.BaseSimulationID), services.ErrSimulationRunning)

	deadline := time.After(2 * time.Second)
	for {
		select {
		case raw := <-events:
			step, ok := raw.(services.SimulationStepEvent)
			if !ok {
				continue
			}
			require.Equal(t, services.BaseSimulationID, step.SimulationID)
			require.Greater(t, step.Tick, uint64(0))
			require.True(t, svc.Simulations()[0].Running)
			return
		case <-deadline:
			t.Fatal("expected simulation step event from running simulation")
		}
	}
}

func TestSimulationServiceEndToEnd_EmitsSimulationEndedEvent(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{
		RunnerConfig: simulation.ControlledRunnerConfig{TicksPerSecond: 128},
	})
	_, err := svc.CreateBaseSimulation(services.BaseSimulationConfig{Options: safeSimulationOptions(1, 1), UntilTick: 3})
	require.NoError(t, err)

	_, events := svc.Broadcaster().Subscribe()
	require.NoError(t, svc.StartSimulation(services.BaseSimulationID))

	deadline := time.After(2 * time.Second)
	for {
		select {
		case raw := <-events:
			ended, ok := raw.(services.SimulationEndedEvent)
			if !ok {
				continue
			}
			require.Equal(t, services.BaseSimulationID, ended.SimulationID)
			require.Equal(t, uint64(3), ended.Tick)
			return
		case <-deadline:
			t.Fatal("expected simulation ended event")
		}
	}
}

func TestSimulationServiceBranch_BaseReturnsRandomNonBaseID(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{Resolution: 10 * time.Minute})
	_, err := svc.CreateBaseSimulation(services.BaseSimulationConfig{Options: safeSimulationOptions(1, 1)})
	require.NoError(t, err)

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)
	require.NotEmpty(t, branchID)
	require.NotEqual(t, services.BaseSimulationID, branchID)

	baseInfo, err := svc.Simulation(services.BaseSimulationID)
	require.NoError(t, err)

	branchInfo, err := svc.Simulation(branchID)
	require.NoError(t, err)
	require.NotNil(t, branchInfo.ParentID)
	require.Equal(t, services.BaseSimulationID, *branchInfo.ParentID)
	require.NotNil(t, branchInfo.SplitTick)
	require.Equal(t, uint64(0), *branchInfo.SplitTick)
	require.NotNil(t, branchInfo.SplitTimestamp)
	require.Equal(t, baseInfo.Tick, *branchInfo.SplitTick)
	require.Equal(t, baseInfo.Timestamp, *branchInfo.SplitTimestamp)

	list := svc.Simulations()
	require.Len(t, list, 2)
	require.Equal(t, services.BaseSimulationID, list[0].ID)
	require.Equal(t, branchID, list[1].ID)
}

func TestSimulationServiceBranch_PausesSourceAndBranch(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{
		RunnerConfig: simulation.ControlledRunnerConfig{TicksPerSecond: 128},
	})
	_, err := svc.CreateBaseSimulation(services.BaseSimulationConfig{Options: safeSimulationOptions(1, 1)})
	require.NoError(t, err)
	require.NoError(t, svc.StartSimulation(services.BaseSimulationID))

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	baseInfo, err := svc.Simulation(services.BaseSimulationID)
	require.NoError(t, err)
	branchInfo, err := svc.Simulation(branchID)
	require.NoError(t, err)

	require.True(t, baseInfo.Paused)
	require.True(t, branchInfo.Paused)
	require.True(t, baseInfo.Running)
	require.True(t, branchInfo.Running)
}

func TestSimulationServiceBranch_IdleBaseRemainsStartableAfterBranch(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{
		RunnerConfig: simulation.ControlledRunnerConfig{TicksPerSecond: 128},
	})
	_, err := svc.CreateBaseSimulation(services.BaseSimulationConfig{Options: safeSimulationOptions(1, 1)})
	require.NoError(t, err)

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)
	require.NotEmpty(t, branchID)

	baseInfo, err := svc.Simulation(services.BaseSimulationID)
	require.NoError(t, err)
	branchInfo, err := svc.Simulation(branchID)
	require.NoError(t, err)
	require.NotNil(t, branchInfo.ParentID)
	require.Equal(t, services.BaseSimulationID, *branchInfo.ParentID)
	require.NotNil(t, branchInfo.SplitTick)
	require.NotNil(t, branchInfo.SplitTimestamp)
	require.Equal(t, baseInfo.Tick, *branchInfo.SplitTick)
	require.Equal(t, baseInfo.Timestamp, *branchInfo.SplitTimestamp)
	require.False(t, baseInfo.Running)
	require.False(t, baseInfo.Paused)
	require.False(t, branchInfo.Running)
	require.False(t, branchInfo.Paused)

	require.NoError(t, svc.StartSimulation(services.BaseSimulationID))
}

func TestSimulationServiceBranchCreatedEvent_IdleBaseEmitsExactLineagePayload(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{Resolution: 10 * time.Minute})
	_, err := svc.CreateBaseSimulation(services.BaseSimulationConfig{Options: safeSimulationOptions(1, 1)})
	require.NoError(t, err)

	clientID, events := svc.Broadcaster().Subscribe()
	t.Cleanup(func() {
		svc.Broadcaster().Unsubscribe(clientID)
	})

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	event := requireNextBranchCreatedEvent(t, events, time.Second)
	require.Equal(t, services.EventTypeBranchCreated, event.Type)
	require.Equal(t, services.BaseSimulationID, event.SimulationID)
	require.Equal(t, branchID, event.BranchID)
	require.Equal(t, services.BaseSimulationID, event.ParentID)

	branchInfo, err := svc.Simulation(branchID)
	require.NoError(t, err)
	require.NotNil(t, branchInfo.SplitTick)
	require.NotNil(t, branchInfo.SplitTimestamp)
	require.Equal(t, *branchInfo.SplitTick, event.SplitTick)
	require.Equal(t, *branchInfo.SplitTimestamp, event.SplitTimestamp)

	requireNoBranchCreatedEvent(t, events, 150*time.Millisecond)
}

func TestSimulationServiceBranchCreatedEvent_RunningBaseEmitsExactLineagePayload(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{
		RunnerConfig: simulation.ControlledRunnerConfig{TicksPerSecond: 128},
	})
	_, err := svc.CreateBaseSimulation(services.BaseSimulationConfig{Options: safeSimulationOptions(1, 1)})
	require.NoError(t, err)
	require.NoError(t, svc.StartSimulation(services.BaseSimulationID))

	clientID, events := svc.Broadcaster().Subscribe()
	t.Cleanup(func() {
		svc.Broadcaster().Unsubscribe(clientID)
	})

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	event := requireNextBranchCreatedEvent(t, events, time.Second)
	require.Equal(t, services.EventTypeBranchCreated, event.Type)
	require.Equal(t, services.BaseSimulationID, event.SimulationID)
	require.Equal(t, branchID, event.BranchID)
	require.Equal(t, services.BaseSimulationID, event.ParentID)

	branchInfo, err := svc.Simulation(branchID)
	require.NoError(t, err)
	require.NotNil(t, branchInfo.SplitTick)
	require.NotNil(t, branchInfo.SplitTimestamp)
	require.Equal(t, *branchInfo.SplitTick, event.SplitTick)
	require.Equal(t, *branchInfo.SplitTimestamp, event.SplitTimestamp)

	requireNoBranchCreatedEvent(t, events, 150*time.Millisecond)
}

func TestSimulationServiceBranchCreatedEvent_MultipleSequentialBaseBranchesEmitSeparateEvents(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{Resolution: 10 * time.Minute})
	_, err := svc.CreateBaseSimulation(services.BaseSimulationConfig{Options: safeSimulationOptions(1, 1)})
	require.NoError(t, err)

	clientID, events := svc.Broadcaster().Subscribe()
	t.Cleanup(func() {
		svc.Broadcaster().Unsubscribe(clientID)
	})

	firstBranchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	firstEvent := requireNextBranchCreatedEvent(t, events, time.Second)
	require.Equal(t, services.EventTypeBranchCreated, firstEvent.Type)
	require.Equal(t, services.BaseSimulationID, firstEvent.SimulationID)
	require.Equal(t, firstBranchID, firstEvent.BranchID)
	require.Equal(t, services.BaseSimulationID, firstEvent.ParentID)

	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))

	secondBranchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	secondEvent := requireNextBranchCreatedEvent(t, events, time.Second)
	require.Equal(t, services.EventTypeBranchCreated, secondEvent.Type)
	require.Equal(t, services.BaseSimulationID, secondEvent.SimulationID)
	require.Equal(t, secondBranchID, secondEvent.BranchID)
	require.Equal(t, services.BaseSimulationID, secondEvent.ParentID)

	firstInfo, err := svc.Simulation(firstBranchID)
	require.NoError(t, err)
	require.NotNil(t, firstInfo.SplitTick)
	require.NotNil(t, firstInfo.SplitTimestamp)
	require.Equal(t, *firstInfo.SplitTick, firstEvent.SplitTick)
	require.Equal(t, *firstInfo.SplitTimestamp, firstEvent.SplitTimestamp)

	secondInfo, err := svc.Simulation(secondBranchID)
	require.NoError(t, err)
	require.NotNil(t, secondInfo.SplitTick)
	require.NotNil(t, secondInfo.SplitTimestamp)
	require.Equal(t, *secondInfo.SplitTick, secondEvent.SplitTick)
	require.Equal(t, *secondInfo.SplitTimestamp, secondEvent.SplitTimestamp)

	require.NotEqual(t, firstEvent.BranchID, secondEvent.BranchID)
	requireNoBranchCreatedEvent(t, events, 150*time.Millisecond)
}

type branchSimulationWithSourceEvent interface {
	BranchSimulationWithSourceEvent(simulationID string, sourceEvent *services.SourceEvent) (string, error)
}

func branchWithSourceEvent(t *testing.T, svc *services.SimulationService, sourceEvent *services.SourceEvent) string {
	t.Helper()

	brancher, ok := any(svc).(branchSimulationWithSourceEvent)
	require.True(t, ok, "SimulationService must support BranchSimulationWithSourceEvent")

	branchID, err := brancher.BranchSimulationWithSourceEvent(services.BaseSimulationID, sourceEvent)
	require.NoError(t, err)

	return branchID
}

func TestSimulationServiceBranchCreatedEvent_SourceEventPersistsAcrossReadModelAndEvent(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{Resolution: 10 * time.Minute})
	_, err := svc.CreateBaseSimulation(services.BaseSimulationConfig{Options: safeSimulationOptions(1, 1)})
	require.NoError(t, err)

	clientID, events := svc.Broadcaster().Subscribe()
	t.Cleanup(func() {
		svc.Broadcaster().Unsubscribe(clientID)
	})

	sourceEvent := &services.SourceEvent{ID: "evt:clicked:1", Type: "task_committed", Tick: 33}
	branchID := branchWithSourceEvent(t, svc, sourceEvent)

	event := requireNextBranchCreatedEvent(t, events, time.Second)
	require.Equal(t, services.EventTypeBranchCreated, event.Type)
	require.Equal(t, services.BaseSimulationID, event.SimulationID)
	require.Equal(t, branchID, event.BranchID)
	require.Equal(t, services.BaseSimulationID, event.ParentID)
	require.NotNil(t, event.SourceEvent)
	require.Equal(t, sourceEvent.ID, event.SourceEvent.ID)
	require.Equal(t, sourceEvent.Type, event.SourceEvent.Type)
	require.Equal(t, sourceEvent.Tick, event.SourceEvent.Tick)

	branchInfo, err := svc.Simulation(branchID)
	require.NoError(t, err)
	require.NotNil(t, branchInfo.SplitTick)
	require.NotNil(t, branchInfo.SplitTimestamp)
	require.Equal(t, *branchInfo.SplitTick, event.SplitTick)
	require.Equal(t, *branchInfo.SplitTimestamp, event.SplitTimestamp)
	require.NotNil(t, branchInfo.SourceEvent)
	require.Equal(t, sourceEvent.ID, branchInfo.SourceEvent.ID)
	require.Equal(t, sourceEvent.Type, branchInfo.SourceEvent.Type)
	require.Equal(t, sourceEvent.Tick, branchInfo.SourceEvent.Tick)

	requireNoBranchCreatedEvent(t, events, 150*time.Millisecond)
}

func TestSimulationServiceBranchCreatedEvent_OmitsSourceEventForLegacyBranchCreation(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{Resolution: 10 * time.Minute})
	_, err := svc.CreateBaseSimulation(services.BaseSimulationConfig{Options: safeSimulationOptions(1, 1)})
	require.NoError(t, err)

	clientID, events := svc.Broadcaster().Subscribe()
	t.Cleanup(func() {
		svc.Broadcaster().Unsubscribe(clientID)
	})

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	event := requireNextBranchCreatedEvent(t, events, time.Second)
	require.Equal(t, services.EventTypeBranchCreated, event.Type)
	require.Equal(t, services.BaseSimulationID, event.SimulationID)
	require.Equal(t, branchID, event.BranchID)
	require.Equal(t, services.BaseSimulationID, event.ParentID)
	require.Nil(t, event.SourceEvent)

	branchInfo, err := svc.Simulation(branchID)
	require.NoError(t, err)
	require.NotNil(t, branchInfo.SplitTick)
	require.NotNil(t, branchInfo.SplitTimestamp)
	require.Equal(t, *branchInfo.SplitTick, event.SplitTick)
	require.Equal(t, *branchInfo.SplitTimestamp, event.SplitTimestamp)
	require.Nil(t, branchInfo.SourceEvent)

	requireNoBranchCreatedEvent(t, events, 150*time.Millisecond)
}

func TestSimulationServiceBranchCreatedEvent_SourceEventTickDoesNotOverrideSplitCoordinates(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{Resolution: 10 * time.Minute})
	_, err := svc.CreateBaseSimulation(services.BaseSimulationConfig{Options: safeSimulationOptions(1, 1)})
	require.NoError(t, err)
	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))
	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))

	baseInfoBeforeBranch, err := svc.Simulation(services.BaseSimulationID)
	require.NoError(t, err)

	clientID, events := svc.Broadcaster().Subscribe()
	t.Cleanup(func() {
		svc.Broadcaster().Unsubscribe(clientID)
	})

	sourceEvent := &services.SourceEvent{ID: "evt:clicked:2", Type: "threat_targeted", Tick: baseInfoBeforeBranch.Tick + 999}
	branchID := branchWithSourceEvent(t, svc, sourceEvent)

	event := requireNextBranchCreatedEvent(t, events, time.Second)
	require.Equal(t, services.EventTypeBranchCreated, event.Type)
	require.Equal(t, services.BaseSimulationID, event.SimulationID)
	require.Equal(t, branchID, event.BranchID)
	require.Equal(t, services.BaseSimulationID, event.ParentID)
	require.NotNil(t, event.SourceEvent)
	require.Equal(t, sourceEvent.Tick, event.SourceEvent.Tick)
	require.Equal(t, baseInfoBeforeBranch.Tick, event.SplitTick)
	require.Equal(t, baseInfoBeforeBranch.Timestamp, event.SplitTimestamp)

	branchInfo, err := svc.Simulation(branchID)
	require.NoError(t, err)
	require.NotNil(t, branchInfo.SplitTick)
	require.NotNil(t, branchInfo.SplitTimestamp)
	require.NotNil(t, branchInfo.SourceEvent)
	require.Equal(t, sourceEvent.ID, branchInfo.SourceEvent.ID)
	require.Equal(t, sourceEvent.Type, branchInfo.SourceEvent.Type)
	require.Equal(t, sourceEvent.Tick, branchInfo.SourceEvent.Tick)
	require.Equal(t, baseInfoBeforeBranch.Tick, *branchInfo.SplitTick)
	require.Equal(t, baseInfoBeforeBranch.Timestamp, *branchInfo.SplitTimestamp)
	require.NotEqual(t, branchInfo.SourceEvent.Tick, *branchInfo.SplitTick)

	requireNoBranchCreatedEvent(t, events, 150*time.Millisecond)
}

func TestSimulationServiceBranchNoEvent_MissingBaseBranchAttempt(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{Resolution: 10 * time.Minute})
	clientID, events := svc.Broadcaster().Subscribe()
	t.Cleanup(func() {
		svc.Broadcaster().Unsubscribe(clientID)
	})

	_, err := svc.BranchSimulation(services.BaseSimulationID)
	require.ErrorIs(t, err, services.ErrBaseNotFound)
	requireNoBranchCreatedEvent(t, events, 150*time.Millisecond)
}

func TestSimulationServiceBranchNoEvent_NonBaseBranchAttempt(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{Resolution: 10 * time.Minute})
	_, err := svc.CreateBaseSimulation(services.BaseSimulationConfig{Options: safeSimulationOptions(1, 1)})
	require.NoError(t, err)

	clientID, events := svc.Broadcaster().Subscribe()
	t.Cleanup(func() {
		svc.Broadcaster().Unsubscribe(clientID)
	})

	_, err = svc.BranchSimulation("branch-a")
	require.ErrorIs(t, err, services.ErrSimulationNotFound)
	requireNoBranchCreatedEvent(t, events, 150*time.Millisecond)
}

func TestSimulationServiceBranch_MissingBaseFails(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{Resolution: 10 * time.Minute})

	_, err := svc.BranchSimulation(services.BaseSimulationID)
	require.ErrorIs(t, err, services.ErrBaseNotFound)
}

func requireNextBranchCreatedEvent(t *testing.T, events <-chan services.Event, timeout time.Duration) services.BranchCreatedEvent {
	t.Helper()

	deadline := time.After(timeout)
	for {
		select {
		case raw := <-events:
			event, ok := raw.(services.BranchCreatedEvent)
			if !ok {
				continue
			}
			return event
		case <-deadline:
			t.Fatal("expected branch created event")
		}
	}
}

func requireNoBranchCreatedEvent(t *testing.T, events <-chan services.Event, timeout time.Duration) {
	t.Helper()

	deadline := time.After(timeout)
	for {
		select {
		case raw := <-events:
			if event, ok := raw.(services.BranchCreatedEvent); ok {
				t.Fatalf("unexpected branch created event: %+v", event)
			}
		case <-deadline:
			return
		}
	}
}

func TestSimulationServiceBranch_NonBaseSimulationRejectedInV1(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{Resolution: 10 * time.Minute})
	_, err := svc.CreateBaseSimulation(services.BaseSimulationConfig{Options: safeSimulationOptions(1, 1)})
	require.NoError(t, err)

	_, err = svc.BranchSimulation("branch-a")
	require.ErrorIs(t, err, services.ErrSimulationNotFound)
}

func TestSimulationServiceBranch_ReadModelsAccessibleByBranchID(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{Resolution: 10 * time.Minute})
	_, err := svc.CreateBaseSimulation(services.BaseSimulationConfig{Options: safeSimulationOptions(2, 2)})
	require.NoError(t, err)

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	airbases, err := svc.Airbases(branchID)
	require.NoError(t, err)
	require.Len(t, airbases, 2)

	aircrafts, err := svc.Aircrafts(branchID)
	require.NoError(t, err)
	require.Len(t, aircrafts, 2)

	threats, err := svc.Threats(branchID)
	require.NoError(t, err)
	require.NotNil(t, threats)
}

func TestSimulationServiceBranch_EmitsBranchScopedEventsOnly(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{Resolution: time.Second})
	_, err := svc.CreateBaseSimulation(services.BaseSimulationConfig{
		Seed:    [32]byte{9, 9, 9},
		Options: positionTrackingSimulationOptions(2),
	})
	require.NoError(t, err)

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	branchInfo, err := svc.Simulation(branchID)
	require.NoError(t, err)
	require.False(t, branchInfo.Paused)
	require.False(t, branchInfo.Running)

	clientID, rawEvents := svc.Broadcaster().Subscribe()
	t.Cleanup(func() {
		svc.Broadcaster().Unsubscribe(clientID)
	})

	deadline := time.After(2 * time.Second)
	seenBranchStep := false
	require.NoError(t, svc.StepSimulation(branchID))
	for !seenBranchStep {
		select {
		case raw := <-rawEvents:
			step, ok := raw.(services.SimulationStepEvent)
			if !ok {
				continue
			}
			require.Equal(t, branchID, step.SimulationID)
			seenBranchStep = true
		case <-deadline:
			t.Fatal("expected branch-scoped step event")
		}
	}

}

func TestSimulationServiceBranch_DeterministicParityAfterEquivalentAdvancement(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{Resolution: time.Second})
	_, err := svc.CreateBaseSimulation(services.BaseSimulationConfig{
		Seed:    [32]byte{4, 5, 6},
		Options: positionTrackingSimulationOptions(3),
	})
	require.NoError(t, err)

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	for range 5 {
		require.NoError(t, svc.StepSimulation(services.BaseSimulationID))
		require.NoError(t, svc.StepSimulation(branchID))
	}

	baseInfo, err := svc.Simulation(services.BaseSimulationID)
	require.NoError(t, err)
	branchInfo, err := svc.Simulation(branchID)
	require.NoError(t, err)
	require.Equal(t, baseInfo.Tick, branchInfo.Tick)
	require.Equal(t, baseInfo.Timestamp, branchInfo.Timestamp)
	require.NotNil(t, branchInfo.ParentID)
	require.Equal(t, services.BaseSimulationID, *branchInfo.ParentID)
	require.NotNil(t, branchInfo.SplitTick)
	require.NotNil(t, branchInfo.SplitTimestamp)
	require.Equal(t, uint64(0), *branchInfo.SplitTick)
	baseSplitTick := *branchInfo.SplitTick
	baseSplitTimestamp := *branchInfo.SplitTimestamp

	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))
	require.NoError(t, svc.StepSimulation(services.BaseSimulationID))
	require.NoError(t, svc.StepSimulation(branchID))
	require.NoError(t, svc.StepSimulation(branchID))

	branchInfoAfter, err := svc.Simulation(branchID)
	require.NoError(t, err)
	require.NotNil(t, branchInfoAfter.SplitTick)
	require.NotNil(t, branchInfoAfter.SplitTimestamp)
	require.Equal(t, baseSplitTick, *branchInfoAfter.SplitTick)
	require.Equal(t, baseSplitTimestamp, *branchInfoAfter.SplitTimestamp)

	baseAircraft, err := svc.Aircrafts(services.BaseSimulationID)
	require.NoError(t, err)
	branchAircraft, err := svc.Aircrafts(branchID)
	require.NoError(t, err)
	require.Equal(t, baseAircraft, branchAircraft)

	baseThreats, err := svc.Threats(services.BaseSimulationID)
	require.NoError(t, err)
	branchThreats, err := svc.Threats(branchID)
	require.NoError(t, err)
	require.Equal(t, baseThreats, branchThreats)
}

func TestSimulationServiceBranch_DeterministicEventIDsDoNotDuplicateBaseID(t *testing.T) {
	t.Parallel()

	svc := services.NewSimulationService(services.SimulationServiceConfig{Resolution: time.Second})
	_, err := svc.CreateBaseSimulation(services.BaseSimulationConfig{
		Seed:    [32]byte{1, 3, 5},
		Options: positionTrackingSimulationOptions(2),
	})
	require.NoError(t, err)

	branchID, err := svc.BranchSimulation(services.BaseSimulationID)
	require.NoError(t, err)

	clientID, rawEvents := svc.Broadcaster().Subscribe()
	t.Cleanup(func() {
		svc.Broadcaster().Unsubscribe(clientID)
	})

	require.NoError(t, svc.StepSimulation(branchID))

	deadline := time.After(time.Second)
	branchStepCount := 0
	for branchStepCount == 0 {
		select {
		case raw := <-rawEvents:
			step, ok := raw.(services.SimulationStepEvent)
			if !ok {
				continue
			}
			require.NotEqual(t, services.BaseSimulationID, step.SimulationID)
			require.Equal(t, branchID, step.SimulationID)
			branchStepCount++
		case <-deadline:
			t.Fatal("expected branch step event without base simulation id")
		}
	}
	require.Equal(t, 1, branchStepCount)
}

func TestAllAircraftPositionsEventEmitted(t *testing.T) {
	t.Parallel()

	const (
		fleetSize = 3
		steps     = 10
	)

	svc := services.NewSimulationService(services.SimulationServiceConfig{Resolution: time.Second})
	_, err := svc.CreateBaseSimulation(services.BaseSimulationConfig{
		Seed:    [32]byte{6, 7, 8},
		Options: positionTrackingSimulationOptions(uint(fleetSize)),
	})
	require.NoError(t, err)

	sim, ok := svc.Base()
	require.True(t, ok)

	initialAircraft := sim.Aircrafts()
	require.Len(t, initialAircraft, fleetSize)

	initialPositions := make(map[simulation.TailNumber][2]float64, fleetSize)
	for _, aircraft := range initialAircraft {
		initialPositions[aircraft.TailNumber] = [2]float64{aircraft.Position.X, aircraft.Position.Y}
	}

	hookEvents := make(chan simulation.AllAircraftPositionsEvent, steps)
	sim.AddAllAircraftPositionsHook(func(event simulation.AllAircraftPositionsEvent) {
		hookEvents <- event
	})

	clientID, rawEvents := svc.Broadcaster().Subscribe()
	t.Cleanup(func() {
		svc.Broadcaster().Unsubscribe(clientID)
	})

	sawMovement := false
	broadcastCount := 0
	for i := range steps {
		sim.Step()

		select {
		case event := <-hookEvents:
			require.Equal(t, uint64(i+1), event.Tick)
			require.Len(t, event.Positions, fleetSize)
			for _, snapshot := range event.Positions {
				initial, ok := initialPositions[snapshot.TailNumber]
				require.True(t, ok)
				if [2]float64{snapshot.Position.X, snapshot.Position.Y} != initial {
					sawMovement = true
				}
			}
		case <-time.After(time.Second):
			t.Fatalf("expected all aircraft positions hook event %d", i+1)
		}

		deadline := time.After(time.Second)
		for {
			select {
			case raw := <-rawEvents:
				event, ok := raw.(services.AllAircraftPositionsEvent)
				if !ok {
					continue
				}
				require.Equal(t, services.EventTypeAllAircraftPositions, event.Type)
				require.Equal(t, services.BaseSimulationID, event.SimulationID)
				require.Equal(t, uint64(i+1), event.Tick)
				require.Len(t, event.Positions, fleetSize)
				broadcastCount++
				goto nextStep
			case <-deadline:
				t.Fatalf("expected all aircraft positions broadcast event %d", i+1)
			}
		}
	nextStep:
	}
	require.Equal(t, steps, broadcastCount)

	select {
	case event := <-hookEvents:
		t.Fatalf("unexpected extra all aircraft positions hook event at tick %d", event.Tick)
	default:
	}
	require.True(t, sawMovement)

	for {
		select {
		case raw := <-rawEvents:
			event, ok := raw.(services.AllAircraftPositionsEvent)
			if !ok {
				continue
			}
			t.Fatalf("unexpected extra all aircraft positions broadcast event at tick %d", event.Tick)
		default:
			return
		}
	}
}

func safeSimulationOptions(numBases, numAircraft uint) *simulation.SimulationOptions {
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

func positionTrackingSimulationOptions(numAircraft uint) *simulation.SimulationOptions {
	lifecycle := simulation.DefaultLifecycleModel()
	lifecycle.Durations.Ready = 0
	return &simulation.SimulationOptions{
		ConstellationOpts: simulation.ConstellationOptions{
			IncludeRegions:    []string{"Blekinge"},
			MinPerRegion:      1,
			MaxPerRegion:      1,
			MaxTotal:          1,
			RegionProbability: prng.New(1, 1),
		},
		FleetOpts: simulation.FleetOptions{
			AircraftMin: numAircraft,
			AircraftMax: numAircraft,
			NeedsMin:    1,
			NeedsMax:    1,
			StateFactory: func(*rand.Rand) simulation.AircraftState {
				return &simulation.ReadyState{}
			},
		},
		ThreatOpts: simulation.ThreatOptions{
			SpawnChance: prng.New(1, 1),
			MaxActive:   numAircraft,
		},
		LifecycleOpts: &lifecycle,
	}
}
