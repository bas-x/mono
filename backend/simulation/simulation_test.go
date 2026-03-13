package simulation

import (
	"math/rand/v2"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/bas-x/basex/assets"
	"github.com/bas-x/basex/geometry"
	"github.com/bas-x/basex/prng"
)

func TestRandSourceCopy(t *testing.T) {
	t.Parallel()
	src := rand.NewChaCha8([32]byte{1, 2, 3})

	_ = src.Uint64()
	_ = src.Uint64()
	_ = src.Uint64()

	srcCopy := *src

	v1 := src.Uint64()
	v2 := srcCopy.Uint64()

	if v1 != v2 {
		t.Errorf("expected both sources to produce the same value, got %d and %d", v1, v2)
	}
}

func TestSimulation_SimpleRun(t *testing.T) {
	t.Parallel()
	ts := New(time.Millisecond, WithEpoch(time.Unix(0, 1)))
	sim := NewSimulator([32]byte{1, 2, 3}, ts)
	runner := NewBasicRunner(BasicRunnerConfig{})
	runner.untilTick = 10

	runner.Run(t.Context(), sim)
	runner.AssertInvariants()
	sim.AssertInvariants()

	require.False(t, runner.active.Load())
}

func TestSimulation_IdenticalStepTagsAfterClone(t *testing.T) {
	t.Parallel()
	t.Run("branch before start", func(t *testing.T) {
		t.Parallel()
		ts := New(time.Millisecond, WithEpoch(time.Unix(0, 1)))
		sim := NewSimulator([32]byte{1, 2, 3}, ts)
		clone := sim.Clone()

		sim.Step()
		clone.Step()

		require.Equal(t, sim.stepTag, clone.stepTag)
	})

	t.Run("branch after start", func(t *testing.T) {
		t.Parallel()
		ts := New(time.Millisecond, WithEpoch(time.Unix(0, 1)))
		sim := NewSimulator([32]byte{1, 2, 3}, ts)
		sim.Step()

		clone := sim.Clone()
		require.Equal(t, sim.stepTag, clone.stepTag)

		sim.Step()
		clone.Step()

		require.Equal(t, sim.stepTag, clone.stepTag)
	})
}

func TestSimulation_RunnerPauseAndBranch(t *testing.T) {
	t.Parallel()
	t.Run("branch while paused", func(t *testing.T) {
		t.Parallel()
		synctest.Test(t, func(t *testing.T) {
			ts := New(time.Millisecond, WithEpoch(time.Unix(0, 1)))
			runnerConfig := ControlledRunnerConfig{TicksPerSecond: 64}

			sim := NewSimulator([32]byte{1, 2, 3}, ts)
			simRunner := NewControlledRunner(runnerConfig)
			go simRunner.Run(t.Context(), sim, 128)

			time.Sleep(100 * time.Millisecond)
			simRunner.Pause()

			clone := sim.Clone()
			cloneRunner := NewControlledRunner(runnerConfig)
			go cloneRunner.Run(t.Context(), clone, 128)

			simRunner.Unpause()
			time.Sleep(3 * time.Second)

			sim.AssertInvariants()
			simRunner.AssertInvariants()
			clone.AssertInvariants()
			cloneRunner.AssertInvariants()

			require.Equal(t, sim.stepTag, clone.stepTag)
			require.False(t, simRunner.active.Load())
			require.False(t, cloneRunner.active.Load())
		})
	})
}

func TestControlledRunner_RunUntilTickParameterStopsAtRequestedTick(t *testing.T) {
	t.Parallel()
	ts := New(time.Millisecond, WithEpoch(time.Unix(0, 1)))
	sim := NewSimulator([32]byte{4, 5, 6}, ts)
	runner := NewControlledRunner(ControlledRunnerConfig{TicksPerSecond: 256})

	runner.Run(t.Context(), sim, 3)

	require.Equal(t, uint64(3), sim.Tick())
	require.False(t, runner.active.Load())
}

func TestSimulationInitAirbases(t *testing.T) {
	t.Parallel()
	ts := New(time.Millisecond, WithEpoch(time.Unix(0, 1)))
	sim := NewSimulator([32]byte{9, 9, 9}, ts)
	options := &SimulationOptions{
		ConstellationOpts: ConstellationOptions{
			IncludeRegions:    []string{"Blekinge"},
			MinPerRegion:      2,
			MaxPerRegion:      3,
			MaxTotal:          5,
			RegionProbability: prng.New(1, 1),
		},
	}

	require.NoError(t, sim.Init(options))
	bases := sim.Airbases()
	require.NotEmpty(t, bases)
	require.LessOrEqual(t, len(bases), 5)

	region := findRegionByName(t, "Blekinge")
	for _, base := range bases {
		require.Equal(t, "Blekinge", base.Region)
		require.Truef(t, pointInsideRegion(base.Location, region), "base %+v not inside region", base)
	}
}

func TestSimulationInitFleet(t *testing.T) {
	t.Parallel()
	ts := New(time.Millisecond, WithEpoch(time.Unix(0, 1)))
	sim := NewSimulator([32]byte{7, 7, 7}, ts)
	opts := &SimulationOptions{
		FleetOpts: FleetOptions{
			AircraftMin:    3,
			AircraftMax:    3,
			NeedsMin:       1,
			NeedsMax:       2,
			NeedsPool:      []NeedType{NeedFuel, NeedMunitions, NeedRepairs},
			SeverityMin:    30,
			SeverityMax:    90,
			BlockingChance: prng.New(1, 2),
		},
	}

	require.NoError(t, sim.Init(opts))
	aircrafts := sim.Aircrafts()
	require.Len(t, aircrafts, 3)

	seen := make(map[TailNumber]struct{}, len(aircrafts))
	for _, aircraft := range aircrafts {
		aircraft.AssertInvariants()
		require.NotEmpty(t, aircraft.State.Name())
		require.Equal(t, "Ready", aircraft.State.Name())
		require.NotEmpty(t, aircraft.Needs)
		require.LessOrEqual(t, len(aircraft.Needs), 2)
		require.GreaterOrEqual(t, len(aircraft.Needs), 1)
		require.NotContains(t, seen, aircraft.TailNumber)
		seen[aircraft.TailNumber] = struct{}{}
	}
}

func TestSimulationInitFleetInitializesAircraftPositionsFromAirbases(t *testing.T) {
	t.Parallel()
	ts := New(time.Millisecond, WithEpoch(time.Unix(0, 1)))
	sim := NewSimulator([32]byte{7, 7, 7}, ts)
	opts := &SimulationOptions{
		ConstellationOpts: ConstellationOptions{
			IncludeRegions:    []string{"Blekinge"},
			MinPerRegion:      1,
			MaxPerRegion:      1,
			MaxTotal:          1,
			RegionProbability: prng.New(1, 1),
		},
		FleetOpts: FleetOptions{
			AircraftMin: 2,
			AircraftMax: 2,
			NeedsMin:    0,
			NeedsMax:    0,
		},
	}

	require.NoError(t, sim.Init(opts))
	bases := sim.Airbases()
	require.Len(t, bases, 1)
	require.NotEqual(t, geometry.Point{}, bases[0].Location)

	for _, aircraft := range sim.Aircrafts() {
		require.Equal(t, bases[0].Location, aircraft.Position)
	}
}

func TestAircraftStateTransitions(t *testing.T) {
	t.Parallel()
	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	sim := NewSimulator([32]byte{1}, ts)
	sim.lifecycle = testLifecycleModel()

	sim.constellation.airbases = []Airbase{
		{ID: BaseID{0, 0, 0, 0, 0, 0, 0, 1}},
		{ID: BaseID{0, 0, 0, 0, 0, 0, 0, 2}},
	}
	sim.dispatcher = NewDispatcher(sim.constellation, &RoundRobinAssigner{})

	tail := TailNumber{0, 0, 0, 0, 0, 0, 0, 9}
	sim.fleet = &Fleet{aircrafts: []Aircraft{NewAircraft(tail, &OutboundState{}, nil)}}

	sim.AssertInvariants()

	resolution := sim.ts.Resolution
	stepsFor := func(d time.Duration) int {
		return int(d/resolution) + 1
	}

	current := func() (Aircraft, string) {
		a := sim.Aircrafts()
		require.NotEmpty(t, a)
		return a[0], a[0].State.Name()
	}

	advance := func(steps int) {
		for i := 0; i < steps; i++ {
			sim.Step()
		}
	}

	_, name := current()
	require.Equal(t, "Outbound", name)
	advance(stepsFor(sim.lifecycle.Durations.Outbound))
	_, name = current()
	require.Equal(t, "Engaged", name)
	advance(stepsFor(sim.lifecycle.Durations.Engaged))
	ac, name := current()
	require.Equal(t, "Inbound", name)
	advance(stepsFor(sim.lifecycle.Durations.InboundDecision))
	ac, name = current()
	require.Equal(t, "Committed", name)
	require.True(t, ac.HasAssignment)
	require.Equal(t, BaseID{0, 0, 0, 0, 0, 0, 0, 1}, ac.AssignedBase)
	advance(stepsFor(sim.lifecycle.Durations.CommitApproach))
	_, name = current()
	require.Equal(t, "Servicing", name)
	advance(stepsFor(sim.lifecycle.Durations.Servicing))
	_, name = current()
	require.Equal(t, "Ready", name)

	_, assigned := sim.Dispatcher().AssignmentFor(tail)
	require.False(t, assigned)
}

func TestSimulationLandingOverrideFlow(t *testing.T) {
	t.Parallel()
	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	sim := NewSimulator([32]byte{2}, ts)
	sim.lifecycle = testLifecycleModel()

	baseA := BaseID{0, 0, 0, 0, 0, 0, 0, 1}
	baseB := BaseID{0, 0, 0, 0, 0, 0, 0, 2}
	sim.constellation.airbases = []Airbase{{ID: baseA}, {ID: baseB}}
	sim.dispatcher = NewDispatcher(sim.constellation, &RoundRobinAssigner{})
	sim.bindInternalHooks()

	tailA := TailNumber{0, 0, 0, 0, 0, 0, 0, 3}
	tailB := TailNumber{0, 0, 0, 0, 0, 0, 0, 4}
	sim.fleet = &Fleet{aircrafts: []Aircraft{
		NewAircraft(tailA, &OutboundState{}, nil),
		NewAircraft(tailB, &OutboundState{}, nil),
	}}

	sim.AssertInvariants()

	resolution := sim.ts.Resolution
	stepsFor := func(d time.Duration) int { return int(d/resolution) + 1 }

	advance := func(steps int) {
		for i := 0; i < steps; i++ {
			sim.Step()
		}
	}

	advance(stepsFor(sim.lifecycle.Durations.Outbound))
	advance(stepsFor(sim.lifecycle.Durations.Engaged))

	_, ok := sim.Dispatcher().AssignmentFor(tailA)
	require.False(t, ok)
	_, ok = sim.Dispatcher().AssignmentFor(tailB)
	require.False(t, ok)

	advance(1)
	_, ok = sim.Dispatcher().AssignmentFor(tailA)
	require.True(t, ok)
	_, ok = sim.Dispatcher().AssignmentFor(tailB)
	require.True(t, ok)

	override, err := sim.OverrideLandingAssignment(tailA, baseB)
	require.NoError(t, err)
	require.Equal(t, baseB, override.Base)
	require.Equal(t, AssignmentSourceHuman, override.Source)

	advance(stepsFor(sim.lifecycle.Durations.InboundDecision))

	acs := sim.Aircrafts()
	require.Len(t, acs, 2)
	require.Equal(t, baseB, acs[0].AssignedBase)
	require.True(t, acs[0].HasAssignment)
	require.True(t, acs[1].HasAssignment)
	require.NotEqual(t, BaseID{}, acs[1].AssignedBase)

	advance(stepsFor(sim.lifecycle.Durations.CommitApproach))
	advance(stepsFor(sim.lifecycle.Durations.Servicing))

	acs = sim.Aircrafts()
	require.Equal(t, "Ready", acs[0].State.Name())
	require.Equal(t, "Ready", acs[1].State.Name())

	_, ok = sim.Dispatcher().AssignmentFor(tailA)
	require.False(t, ok)
	_, ok = sim.Dispatcher().AssignmentFor(tailB)
	require.False(t, ok)
}

func TestSimulationLandingOverridePersistsThroughClone(t *testing.T) {
	t.Parallel()
	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	sim := NewSimulator([32]byte{9}, ts)
	sim.lifecycle = testLifecycleModel()

	baseA := BaseID{0, 0, 0, 0, 0, 0, 0, 1}
	baseB := BaseID{0, 0, 0, 0, 0, 0, 0, 2}
	sim.constellation.airbases = []Airbase{{ID: baseA}, {ID: baseB}}
	sim.dispatcher = NewDispatcher(sim.constellation, &RoundRobinAssigner{})
	sim.bindInternalHooks()

	tail := TailNumber{0, 0, 0, 0, 0, 0, 0, 5}
	_, err := sim.RequestLanding(tail)
	require.NoError(t, err)

	override, err := sim.OverrideLandingAssignment(tail, baseB)
	require.NoError(t, err)
	require.Equal(t, baseB, override.Base)
	require.Equal(t, AssignmentSourceHuman, override.Source)

	clone := sim.Clone()
	require.NotNil(t, clone)

	assignment, ok := clone.Dispatcher().AssignmentFor(tail)
	require.True(t, ok)
	require.Equal(t, baseB, assignment.Base)
	require.Equal(t, AssignmentSourceHuman, assignment.Source)
}

func TestSimulationLandingOverrideEmitsHumanAssignmentEvent(t *testing.T) {
	t.Parallel()
	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	sim := NewSimulator([32]byte{10}, ts)
	sim.lifecycle = testLifecycleModel()

	baseA := BaseID{0, 0, 0, 0, 0, 0, 0, 1}
	baseB := BaseID{0, 0, 0, 0, 0, 0, 0, 2}
	sim.constellation.airbases = []Airbase{{ID: baseA}, {ID: baseB}}
	sim.dispatcher = NewDispatcher(sim.constellation, &RoundRobinAssigner{})
	sim.bindInternalHooks()

	tail := TailNumber{0, 0, 0, 0, 0, 0, 0, 6}
	events := make(chan LandingAssignmentEvent, 4)
	sim.AddLandingAssignmentHook(func(event LandingAssignmentEvent) {
		events <- event
	})

	_, err := sim.RequestLanding(tail)
	require.NoError(t, err)
	select {
	case <-events:
	case <-time.After(time.Second):
		t.Fatal("expected initial landing assignment event")
	}

	override, err := sim.OverrideLandingAssignment(tail, baseB)
	require.NoError(t, err)
	require.Equal(t, baseB, override.Base)
	require.Equal(t, AssignmentSourceHuman, override.Source)

	select {
	case event := <-events:
		require.Equal(t, tail, event.TailNumber)
		require.Equal(t, baseB, event.Base)
		require.Equal(t, AssignmentSourceHuman, event.Source)
	case <-time.After(time.Second):
		t.Fatal("expected human override assignment event")
	default:
		t.Fatal("expected human override assignment event")
	}
}

func TestSimulationLandingOverrideExecutesCommittedBaseSelection(t *testing.T) {
	t.Parallel()
	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	sim := NewSimulator([32]byte{11}, ts)
	sim.lifecycle = testLifecycleModel()

	baseA := BaseID{0, 0, 0, 0, 0, 0, 0, 1}
	baseB := BaseID{0, 0, 0, 0, 0, 0, 0, 2}
	sim.constellation.airbases = []Airbase{{ID: baseA}, {ID: baseB}}
	sim.dispatcher = NewDispatcher(sim.constellation, &RoundRobinAssigner{})

	tail := TailNumber{0, 0, 0, 0, 0, 0, 0, 7}
	sim.fleet = &Fleet{aircrafts: []Aircraft{NewAircraft(tail, &OutboundState{}, nil)}}
	sim.AssertInvariants()

	resolution := sim.ts.Resolution
	stepsFor := func(d time.Duration) int { return int(d/resolution) + 1 }
	advance := func(steps int) {
		for i := 0; i < steps; i++ {
			sim.Step()
		}
	}

	advance(stepsFor(sim.lifecycle.Durations.Outbound))
	advance(stepsFor(sim.lifecycle.Durations.Engaged))
	advance(1)

	override, err := sim.OverrideLandingAssignment(tail, baseB)
	require.NoError(t, err)
	require.Equal(t, baseB, override.Base)

	advance(stepsFor(sim.lifecycle.Durations.InboundDecision))

	aircrafts := sim.Aircrafts()
	require.Len(t, aircrafts, 1)
	require.True(t, aircrafts[0].HasAssignment)
	require.Equal(t, baseB, aircrafts[0].AssignedBase)

	advance(stepsFor(sim.lifecycle.Durations.CommitApproach))

	aircrafts = sim.Aircrafts()
	require.Len(t, aircrafts, 1)
	require.Equal(t, "Servicing", aircrafts[0].State.Name())
	require.Equal(t, baseB, aircrafts[0].AssignedBase)
}

func TestSimulation_NeedsDrivenStateTransitions(t *testing.T) {
	t.Parallel()

	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	sim := NewSimulator([32]byte{3}, ts)
	sim.lifecycle = testLifecycleModel()

	sim.constellation.airbases = []Airbase{{ID: BaseID{0, 0, 0, 0, 0, 0, 0, 1}}}
	sim.dispatcher = NewDispatcher(sim.constellation, &RoundRobinAssigner{})
	tail := TailNumber{0, 0, 0, 0, 0, 0, 0, 7}
	sim.fleet = &Fleet{aircrafts: []Aircraft{NewAircraft(tail, &OutboundState{}, []Need{
		{Type: NeedFuel, Severity: 70, RequiredCapability: NeedFuel},
		{Type: NeedMunitions, Severity: 60, RequiredCapability: NeedMunitions},
	})}}

	advance := func(steps int) {
		for range steps {
			sim.Step()
		}
	}

	current := func() Aircraft {
		aircrafts := sim.Aircrafts()
		require.Len(t, aircrafts, 1)
		return aircrafts[0]
	}

	advance(1)
	ac := current()
	require.Equal(t, "Outbound", ac.State.Name())
	require.Equal(t, 70, ac.Needs[0].Severity)
	require.Equal(t, 60, ac.Needs[1].Severity)

	advance(1)
	ac = current()
	require.Equal(t, "Outbound", ac.State.Name())
	require.Equal(t, 75, ac.Needs[0].Severity)
	require.Equal(t, 65, ac.Needs[1].Severity)

	advance(1)
	ac = current()
	require.Equal(t, "Inbound", ac.State.Name())
	require.Equal(t, 80, ac.Needs[0].Severity)
	require.Equal(t, 70, ac.Needs[1].Severity)

	advance(int(sim.lifecycle.Durations.InboundDecision/ts.Resolution) + 1)
	ac = current()
	require.Equal(t, "Committed", ac.State.Name())
	require.True(t, ac.HasAssignment)

	advance(1)
	ac = current()
	require.Equal(t, "Servicing", ac.State.Name())
	servicingStartFuel := ac.Needs[0].Severity
	servicingStartMunitions := ac.Needs[1].Severity
	require.Positive(t, servicingStartFuel)
	require.Positive(t, servicingStartMunitions)

	advance(1)
	ac = current()
	require.Equal(t, "Servicing", ac.State.Name())
	require.Equal(t, servicingStartFuel, ac.Needs[0].Severity)
	require.Equal(t, servicingStartMunitions, ac.Needs[1].Severity)

	advance(1)
	ac = current()
	require.Equal(t, "Servicing", ac.State.Name())
	require.Less(t, ac.Needs[0].Severity, servicingStartFuel)
	require.Less(t, ac.Needs[1].Severity, servicingStartMunitions)

	advance(int(sim.lifecycle.Durations.Servicing/ts.Resolution) + 1)
	ac = current()
	require.Equal(t, "Ready", ac.State.Name())
	for _, need := range ac.Needs {
		require.Zero(t, need.Severity)
	}

	second := sim.Clone()
	secondSteps := []string{}
	firstSteps := []string{}
	for range 2 {
		sim.Step()
		second.Step()
		firstSteps = append(firstSteps, sim.Aircrafts()[0].State.Name())
		secondSteps = append(secondSteps, second.Aircrafts()[0].State.Name())
	}
	require.Equal(t, firstSteps, secondSteps)
}

func TestThreatSpawnDeterministic(t *testing.T) {
	t.Parallel()

	seed := [32]byte{7, 7, 7}
	opts := &SimulationOptions{
		ConstellationOpts: ConstellationOptions{
			IncludeRegions:    []string{"Blekinge"},
			MinPerRegion:      1,
			MaxPerRegion:      1,
			MaxTotal:          1,
			RegionProbability: prng.New(1, 1),
		},
		ThreatOpts: ThreatOptions{
			SpawnChance: prng.New(1, 1),
			MaxActive:   3,
		},
	}

	ts1 := New(time.Second, WithEpoch(time.Unix(0, 1)))
	sim1 := NewSimulator(seed, ts1)
	require.NoError(t, sim1.Init(opts))

	ts2 := New(time.Second, WithEpoch(time.Unix(0, 1)))
	sim2 := NewSimulator(seed, ts2)
	require.NoError(t, sim2.Init(opts))

	for range 3 {
		sim1.Step()
		sim2.Step()
	}

	require.Equal(t, sim1.Threats(), sim2.Threats())
	require.Len(t, sim1.Threats(), 3)
}

func TestReadyStateRedeploysOnThreat(t *testing.T) {
	t.Parallel()

	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	sim := NewSimulator([32]byte{8}, ts)
	lifecycle := testLifecycleModel()
	lifecycle.Durations.Ready = 0
	sim.lifecycle = lifecycle
	sim.constellation.airbases = []Airbase{{ID: BaseID{0, 0, 0, 0, 0, 0, 0, 1}, RegionID: "SE-K", Region: "Blekinge"}}
	sim.dispatcher = NewDispatcher(sim.constellation, &RoundRobinAssigner{})
	sim.threats = &ThreatSet{pending: []Threat{{ID: makeThreatID(1), Position: geometry.Point{X: mapMinX, Y: mapMinY}, CreatedAt: ts.Now(), CreatedTick: 0}}, active: make(map[ThreatID]Threat)}
	sim.fleet = &Fleet{aircrafts: []Aircraft{NewAircraft(TailNumber{9}, &ReadyState{}, []Need{{Type: NeedFuel, Severity: 0, RequiredCapability: NeedFuel}})}}

	for range 4 {
		sim.Step()
	}
	aircrafts := sim.Aircrafts()
	require.Len(t, aircrafts, 1)
	require.Equal(t, "Outbound", aircrafts[0].State.Name())
	require.Len(t, sim.Threats(), 1)
}

func TestSimulationInitDeterministic(t *testing.T) {
	t.Parallel()
	seed := [32]byte{1, 2, 3}
	opts := &SimulationOptions{
		ConstellationOpts: ConstellationOptions{
			IncludeRegions:    []string{"Blekinge", "Gotland"},
			MinPerRegion:      1,
			MaxPerRegion:      2,
			MaxTotal:          6,
			RegionProbability: prng.New(1, 1),
		},
	}

	ts1 := New(time.Millisecond, WithEpoch(time.Unix(0, 1)))
	sim1 := NewSimulator(seed, ts1)
	require.NoError(t, sim1.Init(opts))

	ts2 := New(time.Millisecond, WithEpoch(time.Unix(0, 1)))
	sim2 := NewSimulator(seed, ts2)
	require.NoError(t, sim2.Init(opts))

	require.Equal(t, sim1.Airbases(), sim2.Airbases())
}

func TestSimulationInitRespectsMaxTotal(t *testing.T) {
	t.Parallel()
	ts := New(time.Millisecond, WithEpoch(time.Unix(0, 1)))
	sim := NewSimulator([32]byte{5, 5, 5}, ts)
	opts := &SimulationOptions{
		ConstellationOpts: ConstellationOptions{
			IncludeRegions:    []string{"Blekinge", "Gotland", "Halland"},
			MinPerRegion:      1,
			MaxPerRegion:      5,
			MaxTotal:          3,
			RegionProbability: prng.New(1, 1),
		},
	}
	require.NoError(t, sim.Init(opts))
	require.LessOrEqual(t, len(sim.Airbases()), 3)
}

func pointInsideRegion(point geometry.Point, region assets.Region) bool {
	for _, area := range region.Areas {
		poly := toGeometryPolygon(area)
		if geometry.PointInPolygon(point, poly) {
			return true
		}
	}
	return false
}

func findRegionByName(t *testing.T, name string) assets.Region {
	t.Helper()
	for _, region := range assets.Regions {
		if strings.EqualFold(region.Name, name) {
			return region
		}
	}
	t.Fatalf("region %s not found", name)
	return assets.Region{}
}
