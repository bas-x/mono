package simulation

import "time"

import "github.com/bas-x/basex/geometry"

type OutboundState struct {
	entered             bool
	enteredAt           time.Time
	lastNeedUpdateAt    time.Time
	needUpdateRemainder time.Duration
}

func (o *OutboundState) Name() string { return "Outbound" }

func (o *OutboundState) Step(a *Aircraft, ctx FlightContext) AircraftState {
	now := ctx.Clock.Now()
	if !o.entered {
		o.entered = true
		o.enteredAt = now
		o.lastNeedUpdateAt = now
		return o
	}
	if a.ClaimedThreat != nil && ctx.ActiveThreats != nil && !ctx.ActiveThreats.IsActive(a.ClaimedThreat.ID) {
		a.ClaimedThreat = nil
		return &InboundState{}
	}
	if a.Speed > 0 {
		target := a.ThreatCentroid
		dist := geometry.Distance(a.Position, target)
		if dist > 0 {
			step := a.Speed * ctx.Clock.Resolution.Seconds()
			if step >= dist {
				a.Position = target
			} else {
				dir := target.Sub(a.Position).Scale(1.0 / dist)
				a.Position = a.Position.Add(dir.Scale(step))
			}
		}
		if geometry.Distance(a.Position, a.ThreatCentroid) <= EngagementRange {
			return &EngagedState{}
		}
	}
	a.ApplyNeedPhase(o.consumeElapsed(now), NeedPhaseOutbound, ctx.Lifecycle, nil)
	if NeedsThresholdReached(a.Needs, ctx.Lifecycle.ReturnThreshold) {
		return &InboundState{}
	}
	if now.Sub(o.enteredAt) >= ctx.Lifecycle.Durations.Outbound {
		return &EngagedState{}
	}
	return o
}

func (o *OutboundState) Clone() AircraftState {
	if o == nil {
		return &OutboundState{}
	}
	clone := *o
	return &clone
}

func (o *OutboundState) consumeElapsed(now time.Time) time.Duration {
	return consumeElapsed(&o.lastNeedUpdateAt, &o.needUpdateRemainder, now)
}
