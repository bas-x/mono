package simulation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bas-x/basex/prng"
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
	sim.lifecycle = testLifecycleModel()
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

func TestSimulationHooks_ThreatEvents(t *testing.T) {
	t.Parallel()

	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	sim := NewSimulator([32]byte{2}, ts)
	sim.lifecycle = testLifecycleModel()
	sim.constellation.airbases = []Airbase{{ID: BaseID{0, 0, 0, 0, 0, 0, 0, 1}, RegionID: "SE-K", Region: "Blekinge"}}
	sim.dispatcher = NewDispatcher(sim.constellation, &RoundRobinAssigner{})
	sim.bindInternalHooks()
	sim.threatOpts = ThreatOptions{SpawnChance: prng.New(1, 1), MaxActive: 1}
	sim.fleet = &Fleet{aircrafts: []Aircraft{NewAircraft(TailNumber{9}, &ReadyState{}, []Need{{Type: NeedFuel, Severity: 0, RequiredCapability: NeedFuel}})}}

	spawned := make(chan ThreatSpawnedEvent, 1)
	claimed := make(chan ThreatClaimedEvent, 1)
	sim.AddThreatSpawnedHook(func(event ThreatSpawnedEvent) { spawned <- event })
	sim.AddThreatClaimedHook(func(event ThreatClaimedEvent) { claimed <- event })

	sim.Step()

	select {
	case event := <-spawned:
		require.NotEmpty(t, event.Threat.Region)
		require.NotEmpty(t, event.Threat.RegionID)
	case <-time.After(time.Second):
		t.Fatal("expected threat spawned event")
	}

	select {
	case event := <-claimed:
		require.Equal(t, TailNumber{9}, event.TailNumber)
		require.NotEmpty(t, event.Threat.Region)
		require.NotEmpty(t, event.Threat.RegionID)
	case <-time.After(time.Second):
		t.Fatal("expected threat claimed event")
	}
}

func TestAllAircraftPositionsHook(t *testing.T) {
	t.Parallel()

	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	sim := NewSimulator([32]byte{42}, ts)
	sim.lifecycle = testLifecycleModel()
	sim.constellation.airbases = []Airbase{{ID: BaseID{0, 0, 0, 0, 0, 0, 0, 1}}}
	sim.dispatcher = NewDispatcher(sim.constellation, &RoundRobinAssigner{})
	sim.bindInternalHooks()

	// Create a fleet with 2 aircraft
	sim.fleet = &Fleet{aircrafts: []Aircraft{
		NewAircraft(TailNumber{0, 0, 0, 0, 0, 0, 0, 1}, &ReadyState{}, nil),
		NewAircraft(TailNumber{0, 0, 0, 0, 0, 0, 0, 2}, &ReadyState{}, nil),
	}}

	positions := make(chan AllAircraftPositionsEvent, 5)
	sim.AddAllAircraftPositionsHook(func(event AllAircraftPositionsEvent) {
		positions <- event
	})

	// Step 5 times
	for range 5 {
		sim.Step()
	}

	// Verify hook was called exactly 5 times
	for i := 0; i < 5; i++ {
		select {
		case event := <-positions:
			require.Equal(t, uint64(i+1), event.Tick)
			require.Equal(t, 2, len(event.Positions))
			require.Equal(t, TailNumber{0, 0, 0, 0, 0, 0, 0, 1}, event.Positions[0].TailNumber)
			require.Equal(t, TailNumber{0, 0, 0, 0, 0, 0, 0, 2}, event.Positions[1].TailNumber)
		case <-time.After(time.Second):
			t.Fatalf("expected all aircraft positions event %d", i+1)
		}
	}

	// Verify no extra events
	select {
	case <-positions:
		t.Fatal("unexpected extra event")
	case <-time.After(100 * time.Millisecond):
		// Expected: no more events
	}
}

func newHookTestSimulation(t *testing.T) *Simulation {
	t.Helper()

	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	sim := NewSimulator([32]byte{1}, ts)
	sim.lifecycle = testLifecycleModel()
	sim.constellation.airbases = []Airbase{{ID: BaseID{0, 0, 0, 0, 0, 0, 0, 1}}}
	sim.dispatcher = NewDispatcher(sim.constellation, &RoundRobinAssigner{})
	sim.bindInternalHooks()
	sim.fleet = &Fleet{aircrafts: []Aircraft{NewAircraft(TailNumber{0, 0, 0, 0, 0, 0, 0, 9}, &OutboundState{}, nil)}}
	return sim
}
