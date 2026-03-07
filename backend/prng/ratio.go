package prng

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
)

// Ratio represents a rational probability in the closed interval [0, 1].
//
// It mirrors the functionality of TigerBeetle's Ratio helper: the numerator must be less than or
// equal to the denominator, and the denominator must be non-zero. All validation occurs at
// construction time so subsequent users can rely on the invariants.
type Ratio struct {
	numerator   uint64
	denominator uint64
}

var (
	// ErrInvalidRatio is returned when parsing finds an invalid ratio string.
	ErrInvalidRatio = errors.New("invalid ratio")
)

// Zero returns the zero probability ratio (0/1).
func Zero() Ratio {
	return Ratio{
		numerator:   0,
		denominator: 1,
	}
}

// New constructs a Ratio, validating numerator and denominator invariants.
func New(numerator, denominator uint64) Ratio {
	if denominator == 0 {
		panic("prng: ratio denominator is zero")
	}
	if numerator > denominator {
		panic("prng: ratio greater than one")
	}
	return Ratio{
		numerator:   numerator,
		denominator: denominator,
	}
}

// MustParse parses a textual ratio in the format "a/b" or "0" and panics on error.
func MustParse(s string) Ratio {
	r, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return r
}

// Parse converts a textual ratio ("a/b" or "0") into a Ratio.
func Parse(s string) (Ratio, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Ratio{}, fmt.Errorf("%w: empty string", ErrInvalidRatio)
	}
	if s == "0" {
		return Zero(), nil
	}
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return Ratio{}, fmt.Errorf("%w: expected 'a/b', got %q", ErrInvalidRatio, s)
	}
	numerator, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return Ratio{}, fmt.Errorf("%w: invalid numerator: %v", ErrInvalidRatio, err)
	}
	denominator, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return Ratio{}, fmt.Errorf("%w: invalid denominator: %v", ErrInvalidRatio, err)
	}
	if denominator == 0 {
		return Ratio{}, fmt.Errorf("%w: denominator is zero", ErrInvalidRatio)
	}
	if numerator > denominator {
		return Ratio{}, fmt.Errorf("%w: numerator greater than denominator", ErrInvalidRatio)
	}
	return Ratio{
		numerator:   numerator,
		denominator: denominator,
	}, nil
}

// Numerator returns the numerator component.
func (r Ratio) Numerator() uint64 {
	return r.numerator
}

// Denominator returns the denominator component.
func (r Ratio) Denominator() uint64 {
	return r.denominator
}

// IsZero reports whether the ratio equals zero.
func (r Ratio) IsZero() bool {
	return r.numerator == 0
}

// IsOne reports whether the ratio equals one.
func (r Ratio) IsOne() bool {
	return r.numerator == r.denominator
}

// String formats the ratio as either "0" or "numerator/denominator".
func (r Ratio) String() string {
	if r.numerator == 0 {
		return "0"
	}
	return fmt.Sprintf("%d/%d", r.numerator, r.denominator)
}

// Chance returns true with the probability represented by Ratio using the provided RNG.
//
// The RNG must yield deterministic sequences when seeded identically so cloned simulations remain
// reproducible.
func Chance(rng *rand.Rand, r Ratio) bool {
	if r.denominator == 0 {
		panic("prng: zero denominator")
	}
	if r.numerator == 0 {
		return false
	}
	if r.numerator == r.denominator {
		return true
	}
	return rng.Uint64N(r.denominator) < r.numerator
}
