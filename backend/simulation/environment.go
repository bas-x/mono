package simulation

import (
	"math/rand/v2"
	"time"

	"github.com/bas-x/basex/assert"
)

type Clock interface {
	Now() time.Time
}

type Environment struct {
	src   *rand.ChaCha8
	rnd   *rand.Rand
	clock Clock
}

func (e *Environment) AssertInvariants() {
	assert.NotNil(e, "environment")
	assert.NotNil(e.src, "rand source")
	assert.NotNil(e.rnd, "rand")
	assert.NotNil(e.clock, "clock")
}

func (e *Environment) Clock() Clock {
	return e.clock
}

func (e *Environment) Rand() *rand.Rand {
	return e.rnd
}

func (e *Environment) Clone(clock Clock) *Environment {
	assert.NotNil(clock, "clock")
	e.AssertInvariants()

	srcCopy := *e.src
	return &Environment{
		src:   &srcCopy,
		rnd:   rand.New(&srcCopy),
		clock: clock,
	}
}
