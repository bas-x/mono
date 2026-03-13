package simulation

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/bas-x/basex/assert"
)

type BasicRunner struct {
	untilTick int64
	pause     chan struct{}
	unpause   chan struct{}
	active    atomic.Bool
}

type BasicRunnerConfig struct {
	UntilTick int64
}

func NewBasicRunner(config BasicRunnerConfig) *BasicRunner {
	return &BasicRunner{
		untilTick: config.UntilTick,
		pause:     make(chan struct{}),
		unpause:   make(chan struct{}),
		active:    atomic.Bool{},
	}
}

func (r *BasicRunner) AssertInvariants() {
	assert.NotNil(r.pause, "pause chan")
	assert.NotNil(r.unpause, "unpause chan")
}

func (r *BasicRunner) Pause() {
	assert.NotNil(r.pause, "pause chan")
	r.pause <- struct{}{}
}

func (r *BasicRunner) Unpause() {
	assert.NotNil(r.unpause, "unpause chan")
	r.unpause <- struct{}{}
}

func (r *BasicRunner) Run(ctx context.Context, s *Simulation) {
	assert.NotNil(ctx, "ctx")
	r.AssertInvariants()

	s.AssertInvariants()
	defer func() {
		s.AssertInvariants()
	}()

	r.active.Store(true)
	defer func() {
		r.active.Store(false)
	}()
	for r.active.Load() {
		select {
		case <-r.pause:
			select {
			case <-r.unpause:
			case <-ctx.Done():
				return
			}
		case <-ctx.Done():
			return
		default:
		}

		s.Step()

		if r.untilTick == int64(s.ts.ticks) {
			return
		}
	}
}

type ControlledRunner struct {
	ticksPerSecond  uint
	maxCatchUpTicks uint
	pause           chan struct{}
	unpause         chan struct{}
	active          atomic.Bool
}

type ControlledRunnerConfig struct {
	TicksPerSecond  uint
	MaxCatchUpTicks uint
}

func NewControlledRunner(config ControlledRunnerConfig) *ControlledRunner {
	if config.TicksPerSecond == 0 {
		config.TicksPerSecond = 64
	}
	if config.MaxCatchUpTicks == 0 {
		config.MaxCatchUpTicks = 5
	}
	return &ControlledRunner{
		ticksPerSecond:  config.TicksPerSecond,
		maxCatchUpTicks: config.MaxCatchUpTicks,
		pause:           make(chan struct{}),
		unpause:         make(chan struct{}),
		active:          atomic.Bool{},
	}
}

func (r *ControlledRunner) AssertInvariants() {
	assert.True(r.ticksPerSecond > 0, "ticks per second > 0", r.ticksPerSecond)
	assert.NotNil(r.pause, "pause chan")
	assert.NotNil(r.unpause, "unpause chan")
}

func (r *ControlledRunner) Pause() {
	assert.NotNil(r.pause, "pause chan")
	r.pause <- struct{}{}
}

func (r *ControlledRunner) Unpause() {
	assert.NotNil(r.unpause, "unpause chan")
	r.unpause <- struct{}{}
}

func (r *ControlledRunner) Run(ctx context.Context, s *Simulation, runUntilTick int64) {
	assert.NotNil(ctx, "ctx")
	r.AssertInvariants()
	s.AssertInvariants()
	defer func() {
		s.AssertInvariants()
	}()
	untilTick := runUntilTick

	r.active.Store(true)
	defer r.active.Store(false)

	tickDuration := time.Second / time.Duration(r.ticksPerSecond)
	nextTick := time.Now()

	for r.active.Load() {
		select {
		case <-r.pause:
			select {
			case <-r.unpause:
			case <-ctx.Done():
				return
			}
		case <-ctx.Done():
			return
		default:
		}

		now := time.Now()
		if now.Before(nextTick) {
			time.Sleep(nextTick.Sub(now))
			continue
		}

		catchUp := 0
		for !now.Before(nextTick) && catchUp < int(r.maxCatchUpTicks) {
			s.Step()

			if untilTick > 0 && untilTick == int64(s.ts.ticks) {
				return
			}

			nextTick = nextTick.Add(tickDuration)
			now = time.Now()
			catchUp++
		}

		if now.After(nextTick) {
			nextTick = now.Add(tickDuration)
		}
	}
}
