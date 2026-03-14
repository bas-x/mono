package simulation

import (
	"testing"
	"time"

	"github.com/bas-x/basex/geometry"
	"github.com/stretchr/testify/require"
)

func TestServicingState_ResetsNeeds(t *testing.T) {
	t.Parallel()

	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	ctx := FlightContext{Clock: ts, Lifecycle: testLifecycleModel()}
	aircraft := NewAircraft(TailNumber{1}, &ServicingState{}, []Need{
		{Type: NeedFuel, Severity: 80, RequiredCapability: NeedFuel},
		{Type: NeedMunitions, Severity: 35, RequiredCapability: NeedMunitions},
	})

	ts.Tick()
	aircraft.Step(ctx)
	require.Equal(t, "Servicing", aircraft.State.Name())
	require.Equal(t, 80, aircraft.Needs[0].Severity)
	require.Equal(t, 35, aircraft.Needs[1].Severity)

	ts.Tick()
	aircraft.Step(ctx)
	require.Equal(t, "Servicing", aircraft.State.Name())
	require.Less(t, aircraft.Needs[0].Severity, 80)
	require.Less(t, aircraft.Needs[1].Severity, 35)

	for range int(testLifecycleModel().Durations.Servicing / ts.Resolution) {
		ts.Tick()
		aircraft.Step(ctx)
	}

	require.Equal(t, "Ready", aircraft.State.Name())
	for _, need := range aircraft.Needs {
		require.Zero(t, need.Severity)
	}
}

func TestOutboundState_DegradeNeeds(t *testing.T) {
	t.Parallel()

	ts := New(500*time.Millisecond, WithEpoch(time.Unix(0, 1)))
	ctx := FlightContext{Clock: ts, Lifecycle: testLifecycleModel()}
	aircraft := NewAircraft(TailNumber{2}, &OutboundState{}, []Need{
		{Type: NeedFuel, Severity: 20, RequiredCapability: NeedFuel},
		{Type: NeedMunitions, Severity: 40, RequiredCapability: NeedMunitions},
	})

	ts.Tick()
	aircraft.Step(ctx)
	require.Equal(t, "Outbound", aircraft.State.Name())
	require.Equal(t, 20, aircraft.Needs[0].Severity)
	require.Equal(t, 40, aircraft.Needs[1].Severity)

	ts.Tick()
	aircraft.Step(ctx)
	require.Equal(t, "Outbound", aircraft.State.Name())
	require.Greater(t, aircraft.Needs[0].Severity, 20)
	require.Greater(t, aircraft.Needs[1].Severity, 40)

	ts.Tick()
	aircraft.Step(ctx)
	require.Greater(t, aircraft.Needs[0].Severity, 22)
	require.Greater(t, aircraft.Needs[1].Severity, 42)
}

func TestOutboundState_EarlyReturnOnThreshold(t *testing.T) {
	t.Parallel()

	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	ctx := FlightContext{Clock: ts, Lifecycle: testLifecycleModel()}
	aircraft := NewAircraft(TailNumber{3}, &OutboundState{}, []Need{
		{Type: NeedFuel, Severity: 75, RequiredCapability: NeedFuel},
	})

	ts.Tick()
	aircraft.Step(ctx)
	require.Equal(t, "Outbound", aircraft.State.Name())
	require.Equal(t, 75, aircraft.Needs[0].Severity)

	ts.Tick()
	aircraft.Step(ctx)
	require.Equal(t, "Inbound", aircraft.State.Name())
	require.Equal(t, 80, aircraft.Needs[0].Severity)
}

func TestEngagedState_DegradeNeeds(t *testing.T) {
	t.Parallel()

	ts := New(500*time.Millisecond, WithEpoch(time.Unix(0, 1)))
	ctx := FlightContext{Clock: ts, Lifecycle: testLifecycleModel()}
	aircraft := NewAircraft(TailNumber{4}, &EngagedState{}, []Need{
		{Type: NeedFuel, Severity: 20, RequiredCapability: NeedFuel},
		{Type: NeedMunitions, Severity: 40, RequiredCapability: NeedMunitions},
	})

	ts.Tick()
	aircraft.Step(ctx)
	require.Equal(t, "Engaged", aircraft.State.Name())
	require.Equal(t, 20, aircraft.Needs[0].Severity)
	require.Equal(t, 40, aircraft.Needs[1].Severity)

	ts.Tick()
	aircraft.Step(ctx)
	require.Equal(t, "Engaged", aircraft.State.Name())
	require.Greater(t, aircraft.Needs[0].Severity, 20)
	require.Greater(t, aircraft.Needs[1].Severity, 40)

	ts.Tick()
	aircraft.Step(ctx)
	require.Greater(t, aircraft.Needs[0].Severity, 22)
	require.Greater(t, aircraft.Needs[1].Severity, 42)
}

func TestEngagedState_EarlyReturnOnThreshold(t *testing.T) {
	t.Parallel()

	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	ctx := FlightContext{Clock: ts, Lifecycle: testLifecycleModel()}
	aircraft := NewAircraft(TailNumber{5}, &EngagedState{}, []Need{
		{Type: NeedFuel, Severity: 75, RequiredCapability: NeedFuel},
	})

	ts.Tick()
	aircraft.Step(ctx)
	require.Equal(t, "Engaged", aircraft.State.Name())
	require.Equal(t, 75, aircraft.Needs[0].Severity)

	ts.Tick()
	aircraft.Step(ctx)
	require.Equal(t, "Inbound", aircraft.State.Name())
	require.Equal(t, 80, aircraft.Needs[0].Severity)
}

func TestServicingState_UsesCapabilityBasedRecovery(t *testing.T) {
	t.Parallel()

	lifecycle := testLifecycleModel()
	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	fuelFast := Airbase{ID: BaseID{1}, Capabilities: map[NeedType]AirbaseCapability{NeedFuel: {RecoveryMultiplierPermille: 1500}}}
	fuelSlow := Airbase{ID: BaseID{2}, Capabilities: map[NeedType]AirbaseCapability{NeedFuel: {RecoveryMultiplierPermille: 500}}}

	fast := NewAircraft(TailNumber{6}, &ServicingState{}, []Need{{Type: NeedFuel, Severity: 80, RequiredCapability: NeedFuel}})
	slow := NewAircraft(TailNumber{7}, &ServicingState{}, []Need{{Type: NeedFuel, Severity: 80, RequiredCapability: NeedFuel}})
	fast.AssignedBase = fuelFast.ID
	fast.HasAssignment = true
	slow.AssignedBase = fuelSlow.ID
	slow.HasAssignment = true

	ctxFast := FlightContext{Clock: ts, Lifecycle: lifecycle, Airbases: []Airbase{fuelFast}}
	ctxSlow := FlightContext{Clock: ts, Lifecycle: lifecycle, Airbases: []Airbase{fuelSlow}}

	ts.Tick()
	fast.Step(ctxFast)
	slow.Step(ctxSlow)
	ts.Tick()
	fast.Step(ctxFast)
	slow.Step(ctxSlow)

	require.Less(t, fast.Needs[0].Severity, slow.Needs[0].Severity)
}

func TestOutboundMovesTowardThreat(t *testing.T) {
	t.Parallel()

	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	ctx := FlightContext{Clock: ts, Lifecycle: testLifecycleModel()}
	aircraft := NewAircraft(TailNumber{8}, &OutboundState{}, nil)
	aircraft.Position = geometry.Point{X: 0, Y: 0}
	aircraft.ThreatCentroid = geometry.Point{X: 200, Y: 200}
	aircraft.Speed = 100

	initialDistance := geometry.Distance(aircraft.Position, aircraft.ThreatCentroid)

	ts.Tick()
	aircraft.Step(ctx)
	ts.Tick()
	aircraft.Step(ctx)

	require.Less(t, geometry.Distance(aircraft.Position, aircraft.ThreatCentroid), initialDistance)
}

func TestOutboundProximityTransition(t *testing.T) {
	t.Parallel()

	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	now := ts.Now()
	ctx := FlightContext{Clock: ts, Lifecycle: testLifecycleModel()}
	state := &OutboundState{entered: true, enteredAt: now, lastNeedUpdateAt: now}
	aircraft := NewAircraft(TailNumber{9}, state, nil)
	aircraft.Position = geometry.Point{X: 5, Y: 5}
	aircraft.ThreatCentroid = geometry.Point{}
	aircraft.Speed = 100

	next := aircraft.State.Step(&aircraft, ctx)

	require.IsType(t, &EngagedState{}, next)
}

func TestEngagedStopsAtThreat(t *testing.T) {
	t.Parallel()

	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	now := ts.Now()
	ctx := FlightContext{Clock: ts, Lifecycle: testLifecycleModel()}
	state := &EngagedState{entered: true, enteredAt: now, lastNeedUpdateAt: now}
	aircraft := NewAircraft(TailNumber{10}, state, nil)
	aircraft.ThreatCentroid = geometry.Point{X: 100, Y: 100}
	aircraft.Position = geometry.Point{X: 80, Y: 120}

	for range 10 {
		next := aircraft.State.Step(&aircraft, ctx)
		require.IsType(t, &EngagedState{}, next)
		require.Equal(t, aircraft.ThreatCentroid, aircraft.Position)
		aircraft.State = next
	}
}

func TestCommittedProximityLanding(t *testing.T) {
	t.Parallel()

	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	now := ts.Now()
	base := Airbase{ID: BaseID{3}, Location: geometry.Point{X: 0, Y: 0}}
	ctx := FlightContext{Clock: ts, Lifecycle: testLifecycleModel(), Airbases: []Airbase{base}}
	state := &CommittedState{entered: true, enteredAt: now}
	aircraft := NewAircraft(TailNumber{11}, state, nil)
	aircraft.AssignedBase = base.ID
	aircraft.HasAssignment = true
	aircraft.Position = geometry.Point{X: 5, Y: 5}

	next := aircraft.State.Step(&aircraft, ctx)

	require.IsType(t, &ServicingState{}, next)
}
