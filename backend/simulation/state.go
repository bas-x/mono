package simulation

import "time"

type FlightContext struct {
	Clock                 *TimeSim
	Dispatcher            *Dispatcher
	OnAircraftStateChange func(AircraftStateChangeEvent)
}

type AircraftState interface {
	Name() string
	Step(aircraft *Aircraft, ctx FlightContext) AircraftState
	Clone() AircraftState
}

const (
	outboundDuration       = 5 * time.Second
	engagedDuration        = 5 * time.Second
	inboundDecisionDelay   = 3 * time.Second
	commitApproachDuration = 4 * time.Second
	servicingDuration      = 6 * time.Second
	needChangePerSecond    = 5
	needsReturnThreshold   = 80
)

type ReadyState struct {
	entered   bool
	enteredAt time.Time
}

func (r *ReadyState) Name() string { return "Ready" }

func (r *ReadyState) Step(_ *Aircraft, ctx FlightContext) AircraftState {
	if !r.entered {
		r.entered = true
		r.enteredAt = ctx.Clock.Now()
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
	a.DegradeNeeds(o.consumeNeedChange(now))
	if NeedsThresholdReached(a.Needs, needsReturnThreshold) {
		return &InboundState{}
	}
	if now.Sub(o.enteredAt) >= outboundDuration {
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
	a.DegradeNeeds(e.consumeNeedChange(now))
	if NeedsThresholdReached(a.Needs, needsReturnThreshold) {
		return &InboundState{}
	}
	if now.Sub(e.enteredAt) >= engagedDuration {
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
	if i.assigned && now.Sub(i.enteredAt) >= inboundDecisionDelay {
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
	if now.Sub(c.enteredAt) >= commitApproachDuration {
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
	entered       bool
	enteredAt     time.Time
	startingNeeds []Need
}

func (s *ServicingState) Name() string { return "Servicing" }

func (s *ServicingState) Step(a *Aircraft, ctx FlightContext) AircraftState {
	now := ctx.Clock.Now()
	if !s.entered {
		s.entered = true
		s.enteredAt = now
		s.startingNeeds = cloneNeeds(a.Needs)
		return s
	}
	s.restoreAircraftNeeds(a, now)
	if now.Sub(s.enteredAt) >= servicingDuration {
		a.ResetNeeds()
		return &ReadyState{}
	}
	return s
}

func (s *ServicingState) Clone() AircraftState {
	if s == nil {
		return &ServicingState{}
	}
	clone := *s
	clone.startingNeeds = cloneNeeds(s.startingNeeds)
	return &clone
}

func (o *OutboundState) consumeNeedChange(now time.Time) int {
	return consumeNeedChange(&o.lastNeedUpdateAt, &o.needUpdateRemainder, now)
}

func (e *EngagedState) consumeNeedChange(now time.Time) int {
	return consumeNeedChange(&e.lastNeedUpdateAt, &e.needUpdateRemainder, now)
}

func consumeNeedChange(lastUpdatedAt *time.Time, remainder *time.Duration, now time.Time) int {
	if lastUpdatedAt.IsZero() {
		*lastUpdatedAt = now
		return 0
	}
	elapsed := now.Sub(*lastUpdatedAt) + *remainder
	if elapsed <= 0 {
		return 0
	}
	changeInterval := time.Second / needChangePerSecond
	amount := int(elapsed / changeInterval)
	if amount == 0 {
		*remainder = elapsed
		*lastUpdatedAt = now
		return 0
	}
	consumed := time.Duration(amount) * changeInterval
	*remainder = elapsed - consumed
	*lastUpdatedAt = now
	return amount
}

func (s *ServicingState) restoreAircraftNeeds(a *Aircraft, now time.Time) {
	if len(s.startingNeeds) == 0 {
		return
	}
	elapsed := now.Sub(s.enteredAt)
	if elapsed < 0 {
		elapsed = 0
	}
	if elapsed > servicingDuration {
		elapsed = servicingDuration
	}
	for i := range a.Needs {
		if i >= len(s.startingNeeds) {
			a.Needs[i].Severity = 0
			continue
		}
		startingSeverity := s.startingNeeds[i].Severity
		remaining := startingSeverity - int((int64(startingSeverity)*elapsed.Nanoseconds())/servicingDuration.Nanoseconds())
		if remaining < 0 {
			remaining = 0
		}
		a.Needs[i].Severity = remaining
	}
}

func cloneNeeds(needs []Need) []Need {
	if len(needs) == 0 {
		return nil
	}
	cloned := make([]Need, len(needs))
	for i := range needs {
		cloned[i] = needs[i].Clone()
	}
	return cloned
}
