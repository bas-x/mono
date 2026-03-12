package simulation

import (
	"math"
	"time"

	"github.com/bas-x/basex/assets"
	"github.com/bas-x/basex/geometry"
)

const (
	EngagementRange        = 25.0
	LandingRange           = 15.0
	OrbitRadius            = 20.0
	OrbitAngleDeltaPerTick = 0.314
)

func threatRegionCentroid(regionName string) geometry.Point {
	for _, region := range assets.Regions {
		if region.Name == regionName {
			if len(region.Areas) == 0 || len(region.Areas[0]) == 0 {
				return geometry.Point{}
			}
			pts := make([]geometry.Point, len(region.Areas[0]))
			for i, p := range region.Areas[0] {
				pts[i] = geometry.Point{X: p.X, Y: p.Y}
			}
			return geometry.PolygonCentroid(pts)
		}
	}
	return geometry.Point{}
}

type FlightContext struct {
	Clock                 *TimeSim
	Dispatcher            *Dispatcher
	Airbases              []Airbase
	Lifecycle             LifecycleModel
	Threats               *ThreatSet
	OnAircraftStateChange func(AircraftStateChangeEvent)
	OnThreatClaimed       func(ThreatClaimedEvent)
}

type AircraftState interface {
	Name() string
	Step(aircraft *Aircraft, ctx FlightContext) AircraftState
	Clone() AircraftState
}

type ReadyState struct {
	entered   bool
	enteredAt time.Time
}

func (r *ReadyState) Name() string { return "Ready" }

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
	if ctx.Threats != nil {
		if threat, ok := ctx.Threats.ClaimNext(); ok {
			a.HasAssignment = false
			a.ClaimedThreat = &threat
			a.ThreatCentroid = threatRegionCentroid(threat.Region)
			a.ResetNeedRemainders()
			if ctx.OnThreatClaimed != nil {
				ctx.OnThreatClaimed(ThreatClaimedEvent{Threat: threat, TailNumber: a.TailNumber, Timestamp: ctx.Clock.Now()})
			}
			return &OutboundState{}
		}
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

type InboundState struct {
	entered    bool
	enteredAt  time.Time
	assignment LandingAssignment
	assigned   bool
}

func (i *InboundState) Name() string { return "Inbound" }

func (i *InboundState) Step(a *Aircraft, ctx FlightContext) AircraftState {
	now := ctx.Clock.Now()
	if !i.entered {
		i.entered = true
		i.enteredAt = now
	}
	assignment, ok := ctx.Dispatcher.AssignmentFor(a.TailNumber)
	if !ok {
		if got, err := ctx.Dispatcher.RegisterInbound(a.TailNumber); err == nil {
			assignment = got
			ok = true
		}
	}
	if ok {
		i.assigned = true
		i.assignment = assignment
	}
	if i.assigned {
		if base, ok := findAssignedAirbase(ctx.Airbases, i.assignment.Base); ok {
			dist := geometry.Distance(a.Position, base.Location)
			if dist > 0 && a.Speed > 0 {
				step := a.Speed * ctx.Clock.Resolution.Seconds()
				if step >= dist {
					a.Position = base.Location
				} else {
					dir := base.Location.Sub(a.Position).Scale(1.0 / dist)
					a.Position = a.Position.Add(dir.Scale(step))
				}
			}
		}
	}
	if i.assigned && now.Sub(i.enteredAt) >= ctx.Lifecycle.Durations.InboundDecision {
		ctx.Dispatcher.RemoveInbound(a.TailNumber)
		a.AssignedBase = i.assignment.Base
		a.HasAssignment = true
		return &CommittedState{}
	}
	return i
}

func (i *InboundState) Clone() AircraftState {
	if i == nil {
		return &InboundState{}
	}
	clone := *i
	return &clone
}

type CommittedState struct {
	entered   bool
	enteredAt time.Time
}

func (c *CommittedState) Name() string { return "Committed" }

func (c *CommittedState) Step(a *Aircraft, ctx FlightContext) AircraftState {
	now := ctx.Clock.Now()
	if !c.entered {
		c.entered = true
		c.enteredAt = now
	}
	if base, ok := findAssignedAirbase(ctx.Airbases, a.AssignedBase); ok {
		dist := geometry.Distance(a.Position, base.Location)
		if dist > 0 && a.Speed > 0 {
			step := a.Speed * ctx.Clock.Resolution.Seconds()
			if step >= dist {
				a.Position = base.Location
			} else {
				dir := base.Location.Sub(a.Position).Scale(1.0 / dist)
				a.Position = a.Position.Add(dir.Scale(step))
			}
		}
		if geometry.Distance(a.Position, base.Location) <= LandingRange {
			return &ServicingState{}
		}
	}
	if now.Sub(c.enteredAt) >= ctx.Lifecycle.Durations.CommitApproach {
		return &ServicingState{}
	}
	return c
}

func (c *CommittedState) Clone() AircraftState {
	if c == nil {
		return &CommittedState{}
	}
	clone := *c
	return &clone
}

type ServicingState struct {
	entered          bool
	enteredAt        time.Time
	lastNeedUpdateAt time.Time
}

func (s *ServicingState) Name() string { return "Servicing" }

func (s *ServicingState) Step(a *Aircraft, ctx FlightContext) AircraftState {
	now := ctx.Clock.Now()
	if !s.entered {
		s.entered = true
		s.enteredAt = now
		s.lastNeedUpdateAt = now
		if base, ok := findAssignedAirbase(ctx.Airbases, a.AssignedBase); ok {
			a.Position = base.Location
		}
		return s
	}
	if base, ok := findAssignedAirbase(ctx.Airbases, a.AssignedBase); ok {
		a.ApplyNeedPhase(s.consumeElapsed(now), NeedPhaseServicing, ctx.Lifecycle, &base)
	} else {
		a.ApplyNeedPhase(s.consumeElapsed(now), NeedPhaseServicing, ctx.Lifecycle, nil)
	}
	if now.Sub(s.enteredAt) >= ctx.Lifecycle.Durations.Servicing {
		a.ResetNeeds()
		a.ResetNeedRemainders()
		return &ReadyState{}
	}
	return s
}

func (s *ServicingState) Clone() AircraftState {
	if s == nil {
		return &ServicingState{}
	}
	clone := *s
	return &clone
}

func (o *OutboundState) consumeElapsed(now time.Time) time.Duration {
	return consumeElapsed(&o.lastNeedUpdateAt, &o.needUpdateRemainder, now)
}

func (e *EngagedState) consumeElapsed(now time.Time) time.Duration {
	return consumeElapsed(&e.lastNeedUpdateAt, &e.needUpdateRemainder, now)
}

func (s *ServicingState) consumeElapsed(now time.Time) time.Duration {
	var remainder time.Duration
	return consumeElapsed(&s.lastNeedUpdateAt, &remainder, now)
}

func consumeElapsed(lastUpdatedAt *time.Time, remainder *time.Duration, now time.Time) time.Duration {
	if lastUpdatedAt.IsZero() {
		*lastUpdatedAt = now
		return 0
	}
	elapsed := now.Sub(*lastUpdatedAt) + *remainder
	if elapsed <= 0 {
		return 0
	}
	*remainder = 0
	*lastUpdatedAt = now
	return elapsed
}

func findAssignedAirbase(airbases []Airbase, id BaseID) (Airbase, bool) {
	for _, airbase := range airbases {
		if airbase.ID == id {
			return airbase, true
		}
	}
	return Airbase{}, false
}
