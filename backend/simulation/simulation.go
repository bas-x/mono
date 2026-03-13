package simulation

import (
	"math/rand/v2"
	"time"

	"github.com/bas-x/basex/assert"
)

type Simulation struct {
	ts      *TimeSim
	env     *Environment
	stepTag uint64

	constellation *Constellation
	fleet         *Fleet
	dispatcher    *Dispatcher
	threats       *ThreatSet
	threatOpts    ThreatOptions
	lifecycle     LifecycleModel

	aircraftStateChangeHooks  []AircraftStateChangeHook
	landingAssignmentHooks    []LandingAssignmentHook
	simulationStepHooks       []SimulationStepHook
	threatSpawnedHooks        []ThreatSpawnedHook
	threatTargetedHooks       []ThreatTargetedHook
	threatDespawnedHooks      []ThreatDespawnedHook
	allAircraftPositionsHooks []AllAircraftPositionsHook
}

func (s *Simulation) AssertInvariants() {
	assert.NotNil(s, "simulator")
	assert.NotNil(s.ts, "timesim")
	s.ts.AssertInvariants()
	s.env.AssertInvariants()
	assert.NotNil(s.constellation, "constellation")
	s.constellation.AssertInvariants()
	assert.NotNil(s.fleet, "fleet")
	s.fleet.AssertInvariants()
	assert.NotNil(s.dispatcher, "dispatcher")
	s.dispatcher.AssertInvariants()
}

func NewSimulator(seed [32]byte, ts *TimeSim) *Simulation {
	assert.NotNil(ts, "TimeSim")

	rndSrc := rand.NewChaCha8(seed)
	rnd := rand.New(rndSrc)
	constellation := NewConstellation()
	assigner := &RoundRobinAssigner{}
	sim := &Simulation{
		ts: ts,
		env: &Environment{
			src:   rndSrc,
			rnd:   rnd,
			clock: ts,
		},
		constellation:             constellation,
		fleet:                     NewFleet(),
		dispatcher:                NewDispatcher(constellation, assigner),
		threats:                   NewThreatSet(),
		lifecycle:                 DefaultLifecycleModel(),
		aircraftStateChangeHooks:  make([]AircraftStateChangeHook, 0),
		landingAssignmentHooks:    make([]LandingAssignmentHook, 0),
		simulationStepHooks:       make([]SimulationStepHook, 0),
		threatSpawnedHooks:        make([]ThreatSpawnedHook, 0),
		threatTargetedHooks:       make([]ThreatTargetedHook, 0),
		threatDespawnedHooks:      make([]ThreatDespawnedHook, 0),
		allAircraftPositionsHooks: make([]AllAircraftPositionsHook, 0),
	}
	sim.bindInternalHooks()
	return sim
}

type SimulationOptions struct {
	ConstellationOpts ConstellationOptions
	FleetOpts         FleetOptions
	ThreatOpts        ThreatOptions
	LifecycleOpts     *LifecycleModel
}

func (s *Simulation) Init(config *SimulationOptions) error {
	var opts *ConstellationOptions
	var fleetOpts *FleetOptions
	if config != nil {
		copyOpts := config.ConstellationOpts
		opts = &copyOpts
		copyFleet := config.FleetOpts
		fleetOpts = &copyFleet
		s.threatOpts = config.ThreatOpts
		if config.LifecycleOpts != nil {
			s.lifecycle = *config.LifecycleOpts
		}
	}
	if err := s.constellation.Init(s.env, opts); err != nil {
		return err
	}
	if err := s.fleet.Init(s.env, fleetOpts); err != nil {
		return err
	}
	s.AssertInvariants()
	return nil
}

func (s *Simulation) Step() {
	s.env.AssertInvariants()
	s.stepTag = s.env.Rand().Uint64()
	s.ts.Tick()

	if spawned, ok := s.threats.TrySpawnEdge(s.env, s.threatOpts, s.ts.Ticks()); ok {
		safeInvoke(s.threatSpawnedHooks, ThreatSpawnedEvent{Threat: spawned, Timestamp: s.ts.Now()})
	}
	despawned := s.threats.DespawnActive(s.ts.Ticks(), s.threatOpts.MaxActiveTicks)
	for _, t := range despawned {
		safeInvoke(s.threatDespawnedHooks, ThreatDespawnedEvent{Threat: t, Timestamp: s.ts.Now()})
	}
	ctx := FlightContext{
		Clock:         s.ts,
		Dispatcher:    s.dispatcher,
		Airbases:      s.constellation.Airbases(),
		Lifecycle:     s.lifecycle,
		Threats:       s.threats,
		ActiveThreats: s.threats,
		OnAircraftStateChange: func(event AircraftStateChangeEvent) {
			safeInvoke(s.aircraftStateChangeHooks, event)
		},
		OnThreatTargeted: func(event ThreatTargetedEvent) {
			safeInvoke(s.threatTargetedHooks, event)
		},
	}
	s.fleet.StepWithContext(ctx)

	aircrafts := s.fleet.Aircrafts()
	snapshots := make([]AircraftPositionSnapshot, len(aircrafts))
	for i, a := range aircrafts {
		snapshots[i] = AircraftPositionSnapshot{
			TailNumber: a.TailNumber,
			Position:   a.Position,
			State:      a.State.Name(),
			Needs:      append([]Need(nil), a.Needs...),
		}
	}
	safeInvoke(s.allAircraftPositionsHooks, AllAircraftPositionsEvent{
		Tick:      s.ts.Ticks(),
		Timestamp: s.ts.Now(),
		Positions: snapshots,
	})

	safeInvoke(s.simulationStepHooks, SimulationStepEvent{Tick: s.ts.Ticks(), Timestamp: s.ts.Now()})
}

// Clone deep copies the simulation. It pauses the
// simulation before doing so.
func (s *Simulation) Clone() *Simulation {
	ts := s.ts.Clone()
	env := s.env.Clone(ts)
	var constellation *Constellation
	if s.constellation != nil {
		constellation = s.constellation.Clone()
	} else {
		constellation = NewConstellation()
	}
	var fleet *Fleet
	if s.fleet != nil {
		fleet = s.fleet.Clone()
	} else {
		fleet = NewFleet()
	}
	var dispatcher *Dispatcher
	if s.dispatcher != nil {
		dispatcher = s.dispatcher.CloneWithConstellation(constellation)
	} else {
		dispatcher = NewDispatcher(constellation, &RoundRobinAssigner{})
	}
	clone := &Simulation{
		ts:                        ts,
		env:                       env,
		stepTag:                   s.stepTag,
		constellation:             constellation,
		fleet:                     fleet,
		dispatcher:                dispatcher,
		threats:                   s.threats.Clone(),
		threatOpts:                s.threatOpts,
		lifecycle:                 s.lifecycle,
		aircraftStateChangeHooks:  append([]AircraftStateChangeHook(nil), s.aircraftStateChangeHooks...),
		landingAssignmentHooks:    append([]LandingAssignmentHook(nil), s.landingAssignmentHooks...),
		simulationStepHooks:       append([]SimulationStepHook(nil), s.simulationStepHooks...),
		threatSpawnedHooks:        append([]ThreatSpawnedHook(nil), s.threatSpawnedHooks...),
		threatTargetedHooks:       append([]ThreatTargetedHook(nil), s.threatTargetedHooks...),
		threatDespawnedHooks:      append([]ThreatDespawnedHook(nil), s.threatDespawnedHooks...),
		allAircraftPositionsHooks: append([]AllAircraftPositionsHook(nil), s.allAircraftPositionsHooks...),
	}
	clone.bindInternalHooks()
	return clone
}

// Airbases returns a shallow copy of the generated airbases slice.
func (s *Simulation) Airbases() []Airbase {
	if s.constellation == nil {
		return nil
	}
	return s.constellation.Airbases()
}

func (s *Simulation) Aircrafts() []Aircraft {
	if s.fleet == nil {
		return nil
	}
	return s.fleet.Aircrafts()
}

func (s *Simulation) Threats() []Threat {
	if s.threats == nil {
		return nil
	}
	return s.threats.Pending()
}

func (s *Simulation) Tick() uint64 {
	if s == nil || s.ts == nil {
		return 0
	}
	return s.ts.Ticks()
}

func (s *Simulation) Now() time.Time {
	if s == nil || s.ts == nil {
		return time.Time{}
	}
	return s.ts.Now()
}

func (s *Simulation) Dispatcher() *Dispatcher {
	return s.dispatcher
}

func (s *Simulation) AddAircraftStateChangeHook(hook AircraftStateChangeHook) {
	assert.NotNil(hook, "aircraft state change hook")
	s.aircraftStateChangeHooks = append(s.aircraftStateChangeHooks, hook)
}

func (s *Simulation) AddLandingAssignmentHook(hook LandingAssignmentHook) {
	assert.NotNil(hook, "landing assignment hook")
	s.landingAssignmentHooks = append(s.landingAssignmentHooks, hook)
}

func (s *Simulation) AddSimulationStepHook(hook SimulationStepHook) {
	assert.NotNil(hook, "simulation step hook")
	s.simulationStepHooks = append(s.simulationStepHooks, hook)
}

func (s *Simulation) AddThreatSpawnedHook(hook ThreatSpawnedHook) {
	assert.NotNil(hook, "threat spawned hook")
	s.threatSpawnedHooks = append(s.threatSpawnedHooks, hook)
}

func (s *Simulation) AddThreatTargetedHook(hook ThreatTargetedHook) {
	assert.NotNil(hook, "threat targeted hook")
	s.threatTargetedHooks = append(s.threatTargetedHooks, hook)
}

func (s *Simulation) AddThreatDespawnedHook(hook ThreatDespawnedHook) {
	assert.NotNil(hook, "threat despawned hook")
	s.threatDespawnedHooks = append(s.threatDespawnedHooks, hook)
}

func (s *Simulation) AddAllAircraftPositionsHook(hook AllAircraftPositionsHook) {
	assert.NotNil(hook, "all aircraft positions hook")
	s.allAircraftPositionsHooks = append(s.allAircraftPositionsHooks, hook)
}

func (s *Simulation) bindInternalHooks() {
	if s.dispatcher == nil {
		return
	}
	s.dispatcher.now = s.ts.Now
	s.dispatcher.onAssignment = func(event LandingAssignmentEvent) {
		safeInvoke(s.landingAssignmentHooks, event)
	}
}

// RequestLanding registers an inbound aircraft and returns its current assignment.
func (s *Simulation) RequestLanding(tail TailNumber) (LandingAssignment, error) {
	return s.dispatcher.RegisterInbound(tail)
}

// OverrideLandingAssignment forces a tail to land at a specific base.
func (s *Simulation) OverrideLandingAssignment(tail TailNumber, base BaseID) (LandingAssignment, error) {
	return s.dispatcher.OverrideAssignment(tail, base)
}

// ClearLandingOverride reverts an override to the algorithmic choice.
func (s *Simulation) ClearLandingOverride(tail TailNumber) (LandingAssignment, error) {
	return s.dispatcher.ClearOverride(tail)
}
