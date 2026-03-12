package geometry

import "math"

// Point represents a 2D coordinate.
type Point struct {
	X float64
	Y float64
}

// Add returns p + q.
func (p Point) Add(q Point) Point {
	return Point{X: p.X + q.X, Y: p.Y + q.Y}
}

// Sub returns p - q.
func (p Point) Sub(q Point) Point {
	return Point{X: p.X - q.X, Y: p.Y - q.Y}
}

// Scale returns p multiplied by scalar s.
func (p Point) Scale(s float64) Point {
	return Point{X: p.X * s, Y: p.Y * s}
}

// Distance returns the Euclidean distance between two points.
func Distance(a, b Point) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// PolygonCentroid returns the arithmetic centroid of a polygon's vertices.
// Returns Point{} for an empty vertex slice.
func PolygonCentroid(vertices []Point) Point {
	if len(vertices) == 0 {
		return Point{}
	}
	var sumX, sumY float64
	for _, v := range vertices {
		sumX += v.X
		sumY += v.Y
	}
	n := float64(len(vertices))
	return Point{X: sumX / n, Y: sumY / n}
}
