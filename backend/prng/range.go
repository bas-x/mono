package prng

import (
	"math/rand/v2"
)

// Unsigned captures the unsigned integer types supported by RangeInclusive.
type Unsigned interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// RangeInclusive returns a uniform random integer in [min, max].
//
// Panics if min > max. The distribution is unbiased as long as the underlying
// RNG produces unbiased Uint64N output (as rand/v2 does for ChaCha).
func RangeInclusive[T Unsigned](rng *rand.Rand, min, max T) T {
	if min > max {
		panic("prng: range min greater than max")
	}
	span := uint64(max - min)
	// span is zero when min == max, so Uint64N receives 1.
	return min + T(rng.Uint64N(span+1))
}
