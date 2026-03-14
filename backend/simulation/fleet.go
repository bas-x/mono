package simulation

import (
	"encoding/binary"
	"math/rand/v2"

	"github.com/bas-x/basex/assert"
	"github.com/bas-x/basex/prng"
)

const (
	MaxFleetSize = 4096
)

type Fleet struct {
	aircrafts []Aircraft
}

type FleetOptions struct {
	AircraftMin    uint
	AircraftMax    uint
	NeedsMin       uint
	NeedsMax       uint
	NeedsPool      []NeedType
	SeverityMin    uint
	SeverityMax    uint
	BlockingChance prng.Ratio
	StateFactory   func(rng *rand.Rand) AircraftState
}

func NewFleet() *Fleet {
	return &Fleet{aircrafts: make([]Aircraft, 0)}
}

func (f *Fleet) AssertInvariants() {
	assert.NotNil(f, "fleet")
	assert.InRange(len(f.aircrafts), 0, MaxFleetSize, "fleet size")
	seen := make(map[TailNumber]struct{}, len(f.aircrafts))
	for i := range f.aircrafts {
		aircraft := &f.aircrafts[i]
		aircraft.AssertInvariants()
		if _, ok := seen[aircraft.TailNumber]; ok {
			assert.Fail("fleet duplicate tail number", aircraft.TailNumber)
		}
		seen[aircraft.TailNumber] = struct{}{}
	}
}

func (f *Fleet) Init(env *Environment, opts *FleetOptions) error {
	assert.NotNil(env, "environment")
	env.AssertInvariants()
	if f.aircrafts == nil {
		f.aircrafts = make([]Aircraft, 0)
	}
	if opts == nil {
		f.aircrafts = f.aircrafts[:0]
		f.AssertInvariants()
		return nil
	}

	normalized := normalizeFleetOptions(*opts)
	rng := env.Rand()
	count := int(prng.RangeInclusive(rng, normalized.AircraftMin, normalized.AircraftMax))
	aircrafts := make([]Aircraft, 0, count)
	seenTailNumbers := make(map[TailNumber]struct{}, count)

	for range count {
		tail := generateTailNumber(rng, seenTailNumbers)
		needs := generateNeeds(rng, normalized)
		state := normalized.StateFactory(rng)
		assert.NotNil(state, "fleet state factory result")
		aircraft := NewAircraft(tail, state, needs)
		aircraft.Speed = aircraftSpeed(tail)
		aircrafts = append(aircrafts, aircraft)
	}

	f.aircrafts = aircrafts
	f.AssertInvariants()
	return nil
}

func (f *Fleet) StepWithContext(ctx FlightContext) {
	f.AssertInvariants()
	for i := range f.aircrafts {
		f.aircrafts[i].Step(ctx)
	}
}

func (f *Fleet) Clone() *Fleet {
	cloned := make([]Aircraft, len(f.aircrafts))
	for i, aircraft := range f.aircrafts {
		cloned[i] = *aircraft.Clone()
	}
	return &Fleet{aircrafts: cloned}
}

func (f *Fleet) Aircrafts() []Aircraft {
	copyOf := make([]Aircraft, len(f.aircrafts))
	copy(copyOf, f.aircrafts)
	return copyOf
}

func normalizeFleetOptions(opts FleetOptions) FleetOptions {
	if opts.AircraftMax == 0 && opts.AircraftMin > 0 {
		opts.AircraftMax = opts.AircraftMin
	}
	if opts.AircraftMax < opts.AircraftMin {
		opts.AircraftMax = opts.AircraftMin
	}
	if opts.AircraftMax > uint(MaxFleetSize) {
		opts.AircraftMax = uint(MaxFleetSize)
	}
	if opts.AircraftMin > opts.AircraftMax {
		opts.AircraftMin = opts.AircraftMax
	}

	if len(opts.NeedsPool) == 0 {
		opts.NeedsPool = append([]NeedType(nil), AllNeedTypes...)
	} else {
		opts.NeedsPool = dedupeNeedTypes(opts.NeedsPool)
		if len(opts.NeedsPool) == 0 {
			opts.NeedsPool = append([]NeedType(nil), AllNeedTypes...)
		}
	}

	if opts.NeedsMax == 0 {
		opts.NeedsMax = opts.NeedsMin
	}
	if opts.NeedsMin > opts.NeedsMax {
		opts.NeedsMax = opts.NeedsMin
	}
	if opts.NeedsMax > uint(len(opts.NeedsPool)) {
		opts.NeedsMax = uint(len(opts.NeedsPool))
	}
	if opts.NeedsMin > opts.NeedsMax {
		opts.NeedsMin = opts.NeedsMax
	}

	if opts.SeverityMax == 0 {
		opts.SeverityMax = 90
	}
	if opts.SeverityMin == 0 {
		opts.SeverityMin = 60
	}
	if opts.SeverityMin > opts.SeverityMax {
		opts.SeverityMin = opts.SeverityMax
	}

	if opts.BlockingChance.Denominator() == 0 {
		opts.BlockingChance = prng.Zero()
	}

	if opts.StateFactory == nil {
		opts.StateFactory = func(_ *rand.Rand) AircraftState { return &ReadyState{} }
	}

	return opts
}

func dedupeNeedTypes(types []NeedType) []NeedType {
	var seen uint64
	result := make([]NeedType, 0, len(types))
	for _, t := range types {
		idx, ok := NeedTypeIndex(t)
		if !ok {
			continue
		}
		mask := uint64(1) << idx
		if seen&mask != 0 {
			continue
		}
		seen |= mask
		result = append(result, t)
	}
	return result
}

func generateTailNumber(rng *rand.Rand, seen map[TailNumber]struct{}) TailNumber {
	for {
		var tn TailNumber
		binary.BigEndian.PutUint64(tn[:], rng.Uint64())
		if _, exists := seen[tn]; exists {
			continue
		}
		seen[tn] = struct{}{}
		return tn
	}
}

func generateNeeds(rng *rand.Rand, opts FleetOptions) []Need {
	count := int(prng.RangeInclusive(rng, opts.NeedsMin, opts.NeedsMax))
	if count == 0 {
		return make([]Need, 0)
	}
	pool := make([]NeedType, len(opts.NeedsPool))
	copy(pool, opts.NeedsPool)
	needs := make([]Need, 0, count)
	for range count {
		if len(pool) == 0 {
			break
		}
		idx := int(prng.RangeInclusive(rng, 0, uint(len(pool)-1)))
		needType := pool[idx]
		pool[idx] = pool[len(pool)-1]
		pool = pool[:len(pool)-1]
		severity := int(prng.RangeInclusive(rng, opts.SeverityMin, opts.SeverityMax))
		blocking := prng.Chance(rng, opts.BlockingChance)
		needs = append(needs, Need{
			Type:               needType,
			Severity:           severity,
			RequiredCapability: needType,
			Blocking:           blocking,
		})
	}
	return needs
}

func aircraftSpeed(tail TailNumber) float64 {
	h := int64(binary.BigEndian.Uint64(tail[:]))
	if h < 0 {
		h = -h
	}
	return 0.8 + float64(h%20)*0.005
}
