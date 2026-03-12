package geometry

import (
	"math"
	"testing"
)

func TestDistance(t *testing.T) {
	tests := []struct {
		name     string
		a        Point
		b        Point
		expected float64
	}{
		{
			name:     "zero distance (same point)",
			a:        Point{X: 0, Y: 0},
			b:        Point{X: 0, Y: 0},
			expected: 0.0,
		},
		{
			name:     "3-4-5 right triangle",
			a:        Point{X: 0, Y: 0},
			b:        Point{X: 3, Y: 4},
			expected: 5.0,
		},
		{
			name:     "symmetry (distance a->b equals b->a)",
			a:        Point{X: 1, Y: 2},
			b:        Point{X: 4, Y: 6},
			expected: 5.0,
		},
		{
			name:     "negative coordinates",
			a:        Point{X: -1, Y: -1},
			b:        Point{X: 2, Y: 3},
			expected: 5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Distance(tt.a, tt.b)
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("Distance(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func TestPolygonCentroid(t *testing.T) {
	tests := []struct {
		name     string
		vertices []Point
		expected Point
	}{
		{
			name:     "empty vertices",
			vertices: []Point{},
			expected: Point{X: 0, Y: 0},
		},
		{
			name:     "nil vertices",
			vertices: nil,
			expected: Point{X: 0, Y: 0},
		},
		{
			name: "square centroid",
			vertices: []Point{
				{X: 0, Y: 0},
				{X: 2, Y: 0},
				{X: 2, Y: 2},
				{X: 0, Y: 2},
			},
			expected: Point{X: 1, Y: 1},
		},
		{
			name: "triangle centroid",
			vertices: []Point{
				{X: 0, Y: 0},
				{X: 3, Y: 0},
				{X: 0, Y: 3},
			},
			expected: Point{X: 1, Y: 1},
		},
		{
			name: "single point",
			vertices: []Point{
				{X: 5, Y: 7},
			},
			expected: Point{X: 5, Y: 7},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PolygonCentroid(tt.vertices)
			if math.Abs(result.X-tt.expected.X) > 1e-9 || math.Abs(result.Y-tt.expected.Y) > 1e-9 {
				t.Errorf("PolygonCentroid(%v) = %v, want %v", tt.vertices, result, tt.expected)
			}
		})
	}
}
