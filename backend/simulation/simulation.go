package simulation

import (
	"math/rand/v2"

	"github.com/bas-x/basex/assert"
)

type Simulation struct {
	ts      *TimeSim
	env     *Environment
	stepTag uint64

	constellation *Constellation
	fleet         *Fleet
	dispatcher    *Dispatcher
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
	return &Simulation{
		ts: ts,
		env: &Environment{
			src:   rndSrc,
			rnd:   rnd,
			clock: ts,
		},
		constellation: constellation,
		fleet:         NewFleet(),
		dispatcher:    NewDispatcher(constellation, assigner),
	}
}

type SimulationOptions struct {
	ConstellationOpts ConstellationOptions
	FleetOpts         FleetOptions
}

func (s *Simulation) Init(config *SimulationOptions) error {
	var opts *ConstellationOptions
	var fleetOpts *FleetOptions
	if config != nil {
		copyOpts := config.ConstellationOpts
		opts = &copyOpts
		copyFleet := config.FleetOpts
		fleetOpts = &copyFleet
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

	if s.fleet != nil {
		ctx := FlightContext{
			Clock:      s.ts,
			Dispatcher: s.dispatcher,
		}
		s.fleet.StepWithContext(ctx)
	}
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
	return &Simulation{
		ts:            ts,
		env:           env,
		stepTag:       s.stepTag,
		constellation: constellation,
		fleet:         fleet,
		dispatcher:    dispatcher,
	}
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

func (s *Simulation) Dispatcher() *Dispatcher {
	return s.dispatcher
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
