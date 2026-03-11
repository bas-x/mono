package geometry

import (
	"errors"
	"math"
	"math/rand/v2"

	"github.com/bas-x/basex/assert"
)

// ErrInvalidPolygon indicates that the provided polygon cannot be triangulated (e.g. less than
// three vertices or degenerate).
var ErrInvalidPolygon = errors.New("geometry: invalid polygon")

type triangle struct {
	a    Point
	b    Point
	c    Point
	area float64
}

// RandomPointInPolygon returns a uniformly random point inside the simple polygon described by
// vertices. Vertices must describe a non-self-intersecting polygon in either clockwise or
// counter-clockwise order. The random generator must be deterministic for reproducibility when used
// with simulation cloning.
func RandomPointInPolygon(rng *rand.Rand, vertices []Point) (Point, error) {
	if len(vertices) < 3 {
		return Point{}, ErrInvalidPolygon
	}
	tris, err := triangulate(vertices)
	if err != nil {
		return Point{}, err
	}
	total := 0.0
	for i := range tris {
		if tris[i].area <= 0 {
			return Point{}, ErrInvalidPolygon
		}
		total += tris[i].area
	}
	if total == 0 {
		return Point{}, ErrInvalidPolygon
	}
	threshold := rng.Float64() * total
	accum := 0.0
	var picked triangle
	for _, tri := range tris {
		accum += tri.area
		if threshold <= accum {
			picked = tri
			break
		}
	}
	return samplePointInTriangle(rng, picked.a, picked.b, picked.c), nil
}

func triangulate(vertices []Point) ([]triangle, error) {
	signedArea := polygonSignedArea(vertices)
	if signedArea == 0 {
		return nil, ErrInvalidPolygon
	}
	indices := make([]int, len(vertices))
	if signedArea < 0 {
		// Reverse to ensure counter-clockwise orientation for ear clipping.
		for i := range vertices {
			indices[i] = len(vertices) - 1 - i
		}
	} else {
		for i := range vertices {
			indices[i] = i
		}
	}
	tris := make([]triangle, 0, len(vertices)-2)
	loopGuard := 0
	for len(indices) > 3 {
		loopGuard++
		if loopGuard > len(vertices)*len(vertices) {
			return nil, ErrInvalidPolygon
		}
		earFound := false
		for i := 0; i < len(indices); i++ {
			prevIdx := indices[(i-1+len(indices))%len(indices)]
			currIdx := indices[i]
			nextIdx := indices[(i+1)%len(indices)]
			if !isConvex(vertices[prevIdx], vertices[currIdx], vertices[nextIdx]) {
				continue
			}
			if containsAnyPoint(vertices, indices, prevIdx, currIdx, nextIdx) {
				continue
			}
			area := triangleArea(vertices[prevIdx], vertices[currIdx], vertices[nextIdx])
			if area <= 0 {
				continue
			}
			tris = append(tris, triangle{
				a:    vertices[prevIdx],
				b:    vertices[currIdx],
				c:    vertices[nextIdx],
				area: area,
			})
			indices = append(indices[:i], indices[i+1:]...)
			earFound = true
			break
		}
		if !earFound {
			return nil, ErrInvalidPolygon
		}
	}
	if len(indices) != 3 {
		return nil, ErrInvalidPolygon
	}
	area := triangleArea(vertices[indices[0]], vertices[indices[1]], vertices[indices[2]])
	if area <= 0 {
		return nil, ErrInvalidPolygon
	}
	tris = append(tris, triangle{
		a:    vertices[indices[0]],
		b:    vertices[indices[1]],
		c:    vertices[indices[2]],
		area: area,
	})
	return tris, nil
}

func polygonSignedArea(vertices []Point) float64 {
	sum := 0.0
	for i := range vertices {
		j := (i + 1) % len(vertices)
		sum += vertices[i].X*vertices[j].Y - vertices[j].X*vertices[i].Y
	}
	return sum / 2
}

func triangleArea(a, b, c Point) float64 {
	return math.Abs(cross(b.Sub(a), c.Sub(a))) / 2
}

// PolygonArea returns the absolute area of the polygon described by vertices. The polygon may be
// oriented clockwise or counter-clockwise. Degenerate polygons (fewer than three vertices) yield 0.
func PolygonArea(vertices []Point) float64 {
	if len(vertices) < 3 {
		return 0
	}
	return math.Abs(polygonSignedArea(vertices))
}

func cross(a, b Point) float64 {
	return a.X*b.Y - a.Y*b.X
}

func isConvex(a, b, c Point) bool {
	return cross(b.Sub(a), c.Sub(b)) > 0
}

func containsAnyPoint(vertices []Point, indices []int, ai, bi, ci int) bool {
	triangle := [3]Point{vertices[ai], vertices[bi], vertices[ci]}
	for _, idx := range indices {
		if idx == ai || idx == bi || idx == ci {
			continue
		}
		if pointInTriangle(vertices[idx], triangle[0], triangle[1], triangle[2]) {
			return true
		}
	}
	return false
}

func pointInTriangle(p, a, b, c Point) bool {
	// Barycentric technique.
	den := (b.Y-c.Y)*(a.X-c.X) + (c.X-b.X)*(a.Y-c.Y)
	if den == 0 {
		return false
	}
	l1 := ((b.Y-c.Y)*(p.X-c.X) + (c.X-b.X)*(p.Y-c.Y)) / den
	l2 := ((c.Y-a.Y)*(p.X-c.X) + (a.X-c.X)*(p.Y-c.Y)) / den
	l3 := 1 - l1 - l2
	const eps = 1e-9
	return l1 >= -eps && l2 >= -eps && l3 >= -eps
}

func samplePointInTriangle(rng *rand.Rand, a, b, c Point) Point {
	u := rng.Float64()
	v := rng.Float64()
	s := math.Sqrt(u)
	f := 1 - s
	g := s * (1 - v)
	h := s * v
	return a.Scale(f).Add(b.Scale(g)).Add(c.Scale(h))
}

// SampleFromPolygons returns a uniformly random point sampled from the provided collection of
// polygons. Areas must contain the precalculated area for each corresponding polygon. Any fatal
// error results in a panic to aid deterministic debugging.
func SampleFromPolygons(rng *rand.Rand, polygons [][]Point, maxAttempts uint) (Point, bool) {
	assert.True(len(polygons) > 0, "sample from polygons requires input", "no polygons provided")

	polygonIdx := rng.IntN(len(polygons))
	assert.True(len(polygons[polygonIdx]) >= 3, "polygon must consist of at least 3 points")

	return SampleFromPolygon(rng, polygons[polygonIdx], maxAttempts)
}

// SampleFromPolygon returns a uniformly random point sampled from a single polygon. When no point
// can be sampled, it returns Point{} and false.
func SampleFromPolygon(rng *rand.Rand, polygon []Point, maxAttempts uint) (Point, bool) {
	if len(polygon) < 3 {
		return Point{}, false
	}
	area := PolygonArea(polygon)
	if area <= 0 {
		return Point{}, false
	}

	pt, err := RandomPointInPolygon(rng, polygon)
	if err == nil {
		return pt, true
	}
	if errors.Is(err, ErrInvalidPolygon) {
		if fallback, fbErr := fallbackPointInPolygon(rng, polygon, maxAttempts); fbErr == nil {
			return fallback, true
		}
	}
	return Point{}, false
}

func fallbackPointInPolygon(rng *rand.Rand, poly []Point, maxAttempts uint) (Point, error) {
	if len(poly) < 3 {
		return Point{}, ErrInvalidPolygon
	}
	minX, maxX := poly[0].X, poly[0].X
	minY, maxY := poly[0].Y, poly[0].Y
	for _, pt := range poly[1:] {
		if pt.X < minX {
			minX = pt.X
		}
		if pt.X > maxX {
			maxX = pt.X
		}
		if pt.Y < minY {
			minY = pt.Y
		}
		if pt.Y > maxY {
			maxY = pt.Y
		}
	}

	if minX == maxX || minY == maxY {
		return Point{}, ErrInvalidPolygon
	}

	limit := maxAttempts
	if limit == 0 {
		limit = 64
	}
	for attempt := uint(0); attempt < limit; attempt++ {
		x := rng.Float64()*(maxX-minX) + minX
		y := rng.Float64()*(maxY-minY) + minY
		candidate := Point{X: x, Y: y}
		if PointInPolygon(candidate, poly) {
			return candidate, nil
		}
	}
	return Point{}, errors.New("geometry: fallback sampling exceeded attempts")
}

// PointInPolygon reports whether p lies inside the simple polygon poly. Points on the edge are
// treated as inside, allowing a small tolerance for floating point error.
func PointInPolygon(p Point, poly []Point) bool {
	inside := false
	for i := range poly {
		j := (i + 1) % len(poly)
		xi, yi := poly[i].X, poly[i].Y
		xj, yj := poly[j].X, poly[j].Y
		if ((yi > p.Y) != (yj > p.Y)) && (p.X < (xj-xi)*(p.Y-yi)/(yj-yi+1e-12)+xi) {
			inside = !inside
		}
	}
	return inside
}
