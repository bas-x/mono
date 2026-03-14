package simulation

import (
	"encoding/binary"
	"time"

	"github.com/bas-x/basex/geometry"
)

type ReadyState struct {
	entered   bool
	enteredAt time.Time
}

func (r *ReadyState) Name() string { return "Ready" }

func readyDwellDuration(tail TailNumber, base time.Duration) time.Duration {
	if base <= 0 {
		return 0
	}
	h := binary.BigEndian.Uint64(tail[:])
	permille := int64(h%1501) - 750
	adjusted := int64(base) * (1000 + permille) / 1000
	if adjusted < 0 {
		return 0
	}
	return time.Duration(adjusted)
}

func (r *ReadyState) Step(a *Aircraft, ctx FlightContext) AircraftState {
	if !r.entered {
		r.entered = true
		r.enteredAt = ctx.Clock.Now()
		if a.HasAssignment {
			if base, ok := findAssignedAirbase(ctx.Airbases, a.AssignedBase); ok && a.Position == (geometry.Point{}) {
				a.Position = base.Location
			}
		}
	}
	if dwell := readyDwellDuration(a.TailNumber, ctx.Lifecycle.Durations.Ready); dwell > 0 && ctx.Clock.Now().Sub(r.enteredAt) < dwell {
		return r
	}
	if threat, ok := ctx.Threats.NextTarget(); ok {
		a.HasAssignment = false
		a.ClaimedThreat = &threat
		a.ThreatCentroid = threat.Position
		a.ResetNeedRemainders()
		if ctx.OnThreatTargeted != nil {
			ctx.OnThreatTargeted(ThreatTargetedEvent{Threat: threat, TailNumber: a.TailNumber, Tick: ctx.Clock.Ticks(), Timestamp: ctx.Clock.Now()})
		}
		return &OutboundState{}
	}
	return r
}

func (r *ReadyState) Clone() AircraftState {
	if r == nil {
		return &ReadyState{}
	}
	clone := *r
	return &clone
}
