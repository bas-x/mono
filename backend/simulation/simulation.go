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
}

func (s *Simulation) AssertInvariants() {
	assert.NotNil(s, "simulator")
	assert.NotNil(s.ts, "timesim")
	s.ts.AssertInvariants()
	s.env.AssertInvariants()
	assert.NotNil(s.constellation, "constellation")
	s.constellation.AssertInvariants()
}

func NewSimulator(seed [32]byte, ts *TimeSim) *Simulation {
	assert.NotNil(ts, "TimeSim")

	rndSrc := rand.NewChaCha8(seed)
	rnd := rand.New(rndSrc)
	return &Simulation{
		ts: ts,
		env: &Environment{
			src:   rndSrc,
			rnd:   rnd,
			clock: ts,
		},
		constellation: NewConstellation(),
	}
}

type SimulationOptions struct {
	Airbases ConstellationOptions
}

func (s *Simulation) Init(config *SimulationOptions) error {
	var opts *ConstellationOptions
	if config != nil {
		copyOpts := config.Airbases
		opts = &copyOpts
	}

	if err := s.constellation.Init(s.env, opts); err != nil {
		return err
	}
	s.AssertInvariants()
	return nil
}

func (s *Simulation) step() {
	s.env.AssertInvariants()
	s.stepTag = s.env.Rand().Uint64()
	s.ts.Tick()
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
	return &Simulation{
		ts:            ts,
		env:           env,
		stepTag:       s.stepTag,
		constellation: constellation,
	}
}

// Airbases returns a shallow copy of the generated airbases slice.
func (s *Simulation) Airbases() []Airbase {
	if s.constellation == nil {
		return nil
	}
	return s.constellation.Airbases()
}
