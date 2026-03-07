package geometry

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
