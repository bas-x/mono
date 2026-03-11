package simulation

import (
	"sync"
	"time"

	"github.com/bas-x/basex/assert"
)

// TimeSim advances only when Tick is called; each tick jumps forward by Resolution nanoseconds.
type TimeSim struct {
	Resolution time.Duration

	ticks uint64
	epoch int64

	mutex sync.Mutex
}

type Option func(*TimeSim)

func (ts *TimeSim) AssertInvariants() {
	assert.NotNil(ts, "timesim")
	assert.True(ts.Resolution > 0, "resolution > 0", ts.Resolution)
	assert.True(ts.epoch > 0, "epoch > 0", ts.epoch)
}

func WithEpoch(epoch time.Time) Option {
	return func(ts *TimeSim) {
		ts.epoch = epoch.UnixNano()
	}
}

func New(resolution time.Duration, opts ...Option) *TimeSim {
	ts := &TimeSim{
		Resolution: resolution,
	}
	for _, opt := range opts {
		opt(ts)
	}
	return ts
}

// Tick moves the monotonic clock forward by exactly Resolution nanoseconds.
func (ts *TimeSim) Tick() {
	ts.mutex.Lock()
	ts.ticks++
	ts.mutex.Unlock()
}

// Monotonic returns strictly increasing elapsed time since construction.
func (ts *TimeSim) Monotonic() time.Duration {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()
	return time.Duration(ts.ticks) * ts.Resolution
}

// Realtime returns simulated wall-clock time including configured drift.
func (ts *TimeSim) Realtime() time.Time {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	mono := time.Duration(ts.ticks) * ts.Resolution

	ns := ts.epoch + int64(mono)
	return time.Unix(0, ns)
}

func (ts *TimeSim) Now() time.Time {
	return ts.Realtime()
}

func (ts *TimeSim) Ticks() uint64 {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()
	return ts.ticks
}

func (ts *TimeSim) Clone() *TimeSim {
	return &TimeSim{
		ticks:      ts.ticks,
		epoch:      ts.epoch,
		mutex:      sync.Mutex{},
		Resolution: ts.Resolution,
	}
}
