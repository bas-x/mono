package prng

import (
	"math/rand/v2"
	"testing"
)

func TestRatioParse(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    Ratio
		wantErr bool
	}{
		{"zero", "0", Zero(), false},
		{"simple", "3/4", New(3, 4), false},
		{"trim", " 10/100 ", New(10, 100), false},
		{"bad format", "3", Ratio{}, true},
		{"bad numerator", "x/10", Ratio{}, true},
		{"bad denominator", "3/y", Ratio{}, true},
		{"zero denominator", "3/0", Ratio{}, true},
		{"too big", "5/4", Ratio{}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := Parse(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if r != tc.want {
				t.Fatalf("got %+v, want %+v", r, tc.want)
			}
		})
	}
}

func TestChance(t *testing.T) {
	seed := [32]byte{1, 2, 3}
	rng := rand.New(rand.NewChaCha8(seed))
	r := New(1, 2)

	const trials = 10000
	var hits int
	for i := 0; i < trials; i++ {
		if Chance(rng, r) {
			hits++
		}
	}

	if hits == 0 || hits == trials {
		t.Fatalf("chance produced degenerate results: hits=%d", hits)
	}

	// Determinism: rerun with the same seed and ensure identical hit count.
	rng2 := rand.New(rand.NewChaCha8(seed))
	var hits2 int
	for i := 0; i < trials; i++ {
		if Chance(rng2, r) {
			hits2++
		}
	}
	if hits != hits2 {
		t.Fatalf("expected deterministic sequence, hits=%d hits2=%d", hits, hits2)
	}
}
