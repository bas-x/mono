package simulation

import (
	"math"
	"time"

	"github.com/bas-x/basex/geometry"
)

type EngagedState struct {
	entered             bool
	enteredAt           time.Time
	lastNeedUpdateAt    time.Time
	needUpdateRemainder time.Duration
}

func (e *EngagedState) Name() string { return "Engaged" }

func (e *EngagedState) Step(a *Aircraft, ctx FlightContext) AircraftState {
	now := ctx.Clock.Now()
	if !e.entered {
		e.entered = true
		e.enteredAt = now
		e.lastNeedUpdateAt = now
		return e
	}
	if a.ClaimedThreat != nil && ctx.ActiveThreats != nil && !ctx.ActiveThreats.IsActive(a.ClaimedThreat.ID) {
		a.ClaimedThreat = nil
		return &InboundState{}
	}
	a.OrbitAngle += OrbitAngleDeltaPerTick
	a.Position = geometry.Point{
		X: a.ThreatCentroid.X + OrbitRadius*math.Cos(a.OrbitAngle),
		Y: a.ThreatCentroid.Y + OrbitRadius*math.Sin(a.OrbitAngle),
	}
	a.ApplyNeedPhase(e.consumeElapsed(now), NeedPhaseEngaged, ctx.Lifecycle, nil)
	if NeedsThresholdReached(a.Needs, ctx.Lifecycle.ReturnThreshold) {
		return &InboundState{}
	}
	if now.Sub(e.enteredAt) >= ctx.Lifecycle.Durations.Engaged {
		return &InboundState{}
	}
	return e
}

func (e *EngagedState) Clone() AircraftState {
	if e == nil {
		return &EngagedState{}
	}
	clone := *e
	return &clone
}

func (e *EngagedState) consumeElapsed(now time.Time) time.Duration {
	return consumeElapsed(&e.lastNeedUpdateAt, &e.needUpdateRemainder, now)
}
