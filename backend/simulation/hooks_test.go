package simulation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSimulationHooks_StateChangeRecoversAndContinues(t *testing.T) {
	t.Parallel()

	sim := newHookTestSimulation(t)
	stateChanges := make(chan AircraftStateChangeEvent, 1)

	sim.AddAircraftStateChangeHook(func(AircraftStateChangeEvent) {
		panic("boom")
	})
	sim.AddAircraftStateChangeHook(func(event AircraftStateChangeEvent) {
		stateChanges <- event
	})

	for range 8 {
		sim.Step()
	}

	select {
	case event := <-stateChanges:
		require.Equal(t, "Outbound", event.OldState)
		require.Equal(t, "Engaged", event.NewState)
		require.Equal(t, event.TailNumber, event.Aircraft.TailNumber)
	case <-time.After(time.Second):
		t.Fatal("expected aircraft state change event")
	}
}

func TestSimulationHooks_LandingAssignmentEvent(t *testing.T) {
	t.Parallel()

	sim := newHookTestSimulation(t)
	assignments := make(chan LandingAssignmentEvent, 1)
	sim.AddLandingAssignmentHook(func(event LandingAssignmentEvent) {
		assignments <- event
	})

	tail := sim.fleet.aircrafts[0].TailNumber
	assignment, err := sim.RequestLanding(tail)
	require.NoError(t, err)

	select {
	case event := <-assignments:
		require.Equal(t, tail, event.TailNumber)
		require.Equal(t, assignment.Base, event.Base)
		require.Equal(t, assignment.Source, event.Source)
	case <-time.After(time.Second):
		t.Fatal("expected landing assignment event")
	}
}

func TestSimulationHooks_StepEvent(t *testing.T) {
	t.Parallel()

	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	sim := NewSimulator([32]byte{1}, ts)
	steps := make(chan SimulationStepEvent, 1)
	sim.AddSimulationStepHook(func(event SimulationStepEvent) {
		steps <- event
	})

	sim.Step()

	select {
	case event := <-steps:
		require.Equal(t, uint64(1), event.Tick)
		require.Equal(t, ts.Now(), event.Timestamp)
	case <-time.After(time.Second):
		t.Fatal("expected simulation step event")
	}
}

func newHookTestSimulation(t *testing.T) *Simulation {
	t.Helper()

	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	sim := NewSimulator([32]byte{1}, ts)
	sim.constellation.airbases = []Airbase{{ID: BaseID{0, 0, 0, 0, 0, 0, 0, 1}}}
	sim.dispatcher = NewDispatcher(sim.constellation, &RoundRobinAssigner{})
	sim.bindInternalHooks()
	sim.fleet = &Fleet{aircrafts: []Aircraft{NewAircraft(TailNumber{0, 0, 0, 0, 0, 0, 0, 9}, &OutboundState{}, nil)}}
	return sim
}
