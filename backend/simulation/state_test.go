package simulation

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestServicingState_ResetsNeeds(t *testing.T) {
	t.Parallel()

	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	ctx := FlightContext{Clock: ts}
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
	require.Equal(t, 67, aircraft.Needs[0].Severity)
	require.Equal(t, 30, aircraft.Needs[1].Severity)

	for range int(servicingDuration / ts.Resolution) {
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
	ctx := FlightContext{Clock: ts}
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
	require.Equal(t, 22, aircraft.Needs[0].Severity)
	require.Equal(t, 42, aircraft.Needs[1].Severity)

	ts.Tick()
	aircraft.Step(ctx)
	require.Equal(t, 25, aircraft.Needs[0].Severity)
	require.Equal(t, 45, aircraft.Needs[1].Severity)
}

func TestOutboundState_EarlyReturnOnThreshold(t *testing.T) {
	t.Parallel()

	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	ctx := FlightContext{Clock: ts}
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
	ctx := FlightContext{Clock: ts}
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
	require.Equal(t, 22, aircraft.Needs[0].Severity)
	require.Equal(t, 42, aircraft.Needs[1].Severity)

	ts.Tick()
	aircraft.Step(ctx)
	require.Equal(t, 25, aircraft.Needs[0].Severity)
	require.Equal(t, 45, aircraft.Needs[1].Severity)
}

func TestEngagedState_EarlyReturnOnThreshold(t *testing.T) {
	t.Parallel()

	ts := New(time.Second, WithEpoch(time.Unix(0, 1)))
	ctx := FlightContext{Clock: ts}
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
