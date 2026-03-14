package simulation

import (
	"encoding/binary"
	"time"

	"github.com/bas-x/basex/assert"
	"github.com/bas-x/basex/geometry"
)

type TailNumber [8]byte

type Aircraft struct {
	TailNumber     TailNumber
	Needs          []Need
	NeedRemainders map[NeedType]int64
	State          AircraftState
	AssignedBase   BaseID
	HasAssignment  bool
	Position       geometry.Point // current x/y position in SVG coordinate space
	Speed          float64        // units per simulation-second, deterministic per tail
	OrbitAngle     float64        // radians, used during EngagedState circling
	ClaimedThreat  *Threat        // the threat this aircraft is flying toward; nil when not applicable
	ThreatCentroid geometry.Point // cached centroid of ClaimedThreat's region; zero when no threat
}

func NewAircraft(
	tn TailNumber,
	state AircraftState,
	needs []Need,
) Aircraft {
	assert.NotNil(state, "state")
	if needs == nil {
		needs = make([]Need, 0)
	}
	clonedNeeds := make([]Need, len(needs))
	for i, need := range needs {
		clonedNeeds[i] = need.Clone()
	}
	aircraft := Aircraft{
		TailNumber:     tn,
		State:          state.Clone(),
		Needs:          clonedNeeds,
		NeedRemainders: make(map[NeedType]int64, len(clonedNeeds)),
	}
	aircraft.AssertInvariants()
	return aircraft
}

func (a *Aircraft) Step(ctx FlightContext) {
	a.AssertInvariants()
	oldState := a.State.Name()
	nextState := a.State.Step(a, ctx)
	assert.NotNil(nextState, "next state")
	a.State = nextState
	if oldState != a.State.Name() && ctx.OnAircraftStateChange != nil {
		ctx.OnAircraftStateChange(AircraftStateChangeEvent{
			TailNumber: a.TailNumber,
			OldState:   oldState,
			NewState:   a.State.Name(),
			Aircraft:   *a.Clone(),
			Tick:       ctx.Clock.Ticks(),
			Timestamp:  ctx.Clock.Now(),
		})
	}
}

func (a *Aircraft) ResetNeeds() {
	for i := range a.Needs {
		a.Needs[i].Severity = 0
	}
}

func (a *Aircraft) ResetNeedRemainders() {
	for key := range a.NeedRemainders {
		a.NeedRemainders[key] = 0
	}
}

func (a *Aircraft) DegradeNeeds(amount int) {
	for i := range a.Needs {
		a.Needs[i].Degrade(amount)
	}
}

func (a *Aircraft) RestoreNeeds(amount int) {
	for i := range a.Needs {
		a.Needs[i].Restore(amount)
	}
}

func (a *Aircraft) Clone() *Aircraft {
	clonedNeeds := make([]Need, len(a.Needs))
	for i, need := range a.Needs {
		clonedNeeds[i] = need.Clone()
	}
	return &Aircraft{
		TailNumber:     a.TailNumber,
		State:          a.State.Clone(),
		Needs:          clonedNeeds,
		NeedRemainders: cloneNeedRemainders(a.NeedRemainders),
		AssignedBase:   a.AssignedBase,
		HasAssignment:  a.HasAssignment,
		Position:       a.Position,
		Speed:          a.Speed,
		OrbitAngle:     a.OrbitAngle,
		ThreatCentroid: a.ThreatCentroid,
		ClaimedThreat:  cloneThreat(a.ClaimedThreat),
	}
}

func (a *Aircraft) AssertInvariants() {
	assert.NotNil(a, "aircraft")
	assert.NotNil(a.State, "aircraft state")
	var seen uint64
	for _, need := range a.Needs {
		need.AssertInvariants()

		idx, ok := NeedTypeIndex(need.Type)
		assert.True(ok, "aircraft need type registered", need.Type)

		mask := uint64(1) << idx
		assert.True(seen&mask == 0, "aircraft needs unique", need.Type)

		seen |= mask
	}
	assert.NotNil(a.State, "aircraft state")
}

func (a *Aircraft) ApplyNeedPhase(elapsed time.Duration, phase NeedPhase, lifecycle LifecycleModel, airbase *Airbase) {
	if a.NeedRemainders == nil {
		a.NeedRemainders = make(map[NeedType]int64, len(a.Needs))
	}
	for i := range a.Needs {
		need := &a.Needs[i]
		model := lifecycle.NeedModel(need.Type)
		rate := deterministicNeedRate(a.TailNumber, need.Type, phase, model)
		if phase == NeedPhaseServicing {
			multiplier := int64(1000)
			if airbase != nil {
				multiplier = airbase.RecoveryMultiplier(need.RequiredCapability)
			}
			rate = rate * multiplier / 1000
			a.restoreNeedWithRate(need, elapsed, rate)
			continue
		}
		a.degradeNeedWithRate(need, elapsed, rate)
	}
}

func (a *Aircraft) degradeNeedWithRate(need *Need, elapsed time.Duration, milliPerHour int64) {
	if need == nil || elapsed <= 0 || milliPerHour <= 0 {
		return
	}
	remainder := a.NeedRemainders[need.Type]
	totalMilli := remainder + elapsed.Nanoseconds()*milliPerHour/int64(time.Hour)
	delta := int(totalMilli / 1000)
	a.NeedRemainders[need.Type] = totalMilli % 1000
	need.Degrade(delta)
}

func (a *Aircraft) restoreNeedWithRate(need *Need, elapsed time.Duration, milliPerHour int64) {
	if need == nil || elapsed <= 0 || milliPerHour <= 0 {
		return
	}
	remainder := a.NeedRemainders[need.Type]
	totalMilli := remainder + elapsed.Nanoseconds()*milliPerHour/int64(time.Hour)
	delta := int(totalMilli / 1000)
	a.NeedRemainders[need.Type] = totalMilli % 1000
	need.Restore(delta)
}

func cloneNeedRemainders(input map[NeedType]int64) map[NeedType]int64 {
	if len(input) == 0 {
		return map[NeedType]int64{}
	}
	cloned := make(map[NeedType]int64, len(input))
	for k, v := range input {
		cloned[k] = v
	}
	return cloned
}

func cloneThreat(t *Threat) *Threat {
	if t == nil {
		return nil
	}
	return &Threat{
		ID:          t.ID,
		Position:    t.Position,
		CreatedAt:   t.CreatedAt,
		CreatedTick: t.CreatedTick,
	}
}

func deterministicNeedRate(tail TailNumber, needType NeedType, phase NeedPhase, model NeedRateModel) int64 {
	base := model.RateForPhase(phase)
	if base <= 0 || model.VariancePermille <= 0 {
		return base
	}
	idx, _ := NeedTypeIndex(needType)
	hash := int64(binary.BigEndian.Uint64(tail[:])) + int64(idx*97) + int64(phase*193)
	spread := model.VariancePermille*2 + 1
	offset := (hash % spread) - model.VariancePermille
	adjusted := base * (1000 + offset) / 100
	if adjusted < 1 {
		return 1
	}
	return adjusted
}
