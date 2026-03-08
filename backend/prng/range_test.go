package prng

import (
	"math/rand/v2"
	"testing"
)

func TestRangeInclusiveBounds(t *testing.T) {
	seed := [32]byte{9, 8, 7}
	rng := rand.New(rand.NewChaCha8(seed))

	tests := []struct {
		min, max uint64
	}{
		{0, 0},
		{0, 1},
		{5, 5},
		{3, 9},
		{42, 99},
	}

	for _, tc := range tests {
		for range 1000 {
			v := RangeInclusive[uint64](rng, tc.min, tc.max)
			if v < tc.min || v > tc.max {
				t.Fatalf("value %d outside range [%d,%d]", v, tc.min, tc.max)
			}
		}
	}
}

func TestRangeInclusiveDeterministic(t *testing.T) {
	seed := [32]byte{1, 2, 3}
	rng1 := rand.New(rand.NewChaCha8(seed))
	rng2 := rand.New(rand.NewChaCha8(seed))

	var seq1, seq2 [64]uint64
	for i := range seq1 {
		seq1[i] = RangeInclusive[uint64](rng1, 10, 20)
		seq2[i] = RangeInclusive[uint64](rng2, 10, 20)
	}

	if seq1 != seq2 {
		t.Fatalf("determinism failure: %v != %v", seq1, seq2)
	}
}
