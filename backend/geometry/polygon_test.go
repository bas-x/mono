package geometry

import (
	"math"
	"math/rand/v2"
	"testing"
)

func TestRandomPointInTriangle(t *testing.T) {
	t.Parallel()
	triangle := []Point{{0, 0}, {1, 0}, {0, 1}}
	seed := [32]byte{1, 2, 3}
	rng := rand.New(rand.NewChaCha8(seed))
	for range 100 {
		p, err := RandomPointInPolygon(rng, triangle)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !pointInTriangle(p, triangle[0], triangle[1], triangle[2]) {
			t.Fatalf("point %+v lies outside triangle", p)
		}
	}
}

func TestPolygonArea(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		pts  []Point
		want float64
	}{
		{"degenerate", []Point{{0, 0}, {1, 0}}, 0},
		{"triangle", []Point{{0, 0}, {4, 0}, {0, 3}}, 6},
		{"square", []Point{{0, 0}, {1, 0}, {1, 1}, {0, 1}}, 1},
		{"square cw", []Point{{0, 0}, {0, 1}, {1, 1}, {1, 0}}, 1},
	}

	for _, tc := range cases {
		if got := PolygonArea(tc.pts); math.Abs(got-tc.want) > 1e-9 {
			t.Fatalf("%s: got %f, want %f", tc.name, got, tc.want)
		}
	}
}

func TestRandomPointInConcavePolygon(t *testing.T) {
	t.Parallel()
	poly := []Point{
		{0, 0},
		{4, 0},
		{4, 4},
		{2, 2},
		{0, 4},
	}
	seed := [32]byte{9, 9, 9}
	rng := rand.New(rand.NewChaCha8(seed))
	for range 200 {
		p, err := RandomPointInPolygon(rng, poly)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !pointInPolygon(p, poly) {
			t.Fatalf("point %+v lies outside polygon", p)
		}
	}
}

func TestRandomPointInvalid(t *testing.T) {
	t.Parallel()
	seed := [32]byte{0}
	rng := rand.New(rand.NewChaCha8(seed))
	if _, err := RandomPointInPolygon(rng, nil); err == nil {
		t.Fatalf("expected error for empty polygon")
	}
	if _, err := RandomPointInPolygon(rng, []Point{{0, 0}, {1, 1}}); err == nil {
		t.Fatalf("expected error for degenerate polygon")
	}
}

func TestSampleFromPolygon(t *testing.T) {
	t.Parallel()
	poly := []Point{{0, 0}, {2, 0}, {2, 1}, {0, 1}}
	seed := [32]byte{4, 5, 6}
	rng := rand.New(rand.NewChaCha8(seed))
	for range 100 {
		pt, ok := SampleFromPolygon(rng, poly, 16)
		if !ok {
			t.Fatalf("expected successful sampling")
		}
		if !PointInPolygon(pt, poly) {
			t.Fatalf("sampled point %+v outside polygon", pt)
		}
	}
}

func TestSampleFromPolygonInvalid(t *testing.T) {
	t.Parallel()
	seed := [32]byte{7, 8, 9}
	rng := rand.New(rand.NewChaCha8(seed))
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for invalid polygon")
		}
	}()
	SampleFromPolygon(rng, []Point{{0, 0}, {1, 1}}, 8)
}

func pointInPolygon(p Point, vertices []Point) bool {
	inside := false
	for i := range vertices {
		j := (i + 1) % len(vertices)
		xi, yi := vertices[i].X, vertices[i].Y
		xj, yj := vertices[j].X, vertices[j].Y
		intersect := ((yi > p.Y) != (yj > p.Y)) &&
			(p.X < (xj-xi)*(p.Y-yi)/(yj-yi+1e-12)+xi)
		if intersect {
			inside = !inside
		}
	}
	return inside
}
