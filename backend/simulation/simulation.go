package simulation

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand/v2"
	"sort"
	"strings"

	"github.com/bas-x/basex/assert"
	"github.com/bas-x/basex/assets"
	"github.com/bas-x/basex/geometry"
	"github.com/bas-x/basex/prng"
)

const (
	MaxNumAirbases = 256
)

type Simulation struct {
	ts      *TimeSim
	env     *Environment
	stepTag uint64

	airbases []Airbase
}

func (s *Simulation) AssertInvariants() {
	assert.NotNil(s, "simulator")
	assert.NotNil(s.ts, "timesim")
	s.ts.AssertInvariants()
	s.env.AssertInvariants()
	assert.InRange(len(s.airbases), 0, MaxNumAirbases, "airbases length")

	seen := make(map[BaseID]struct{}, len(s.airbases))
	for _, base := range s.airbases {
		if _, ok := seen[base.ID]; ok {
			panic("duplicate airbase id")
		}
		seen[base.ID] = struct{}{}
	}
}

func NewSimulator(seed [32]byte, ts *TimeSim) *Simulation {
	assert.NotNil(ts, "TimeSim")

	rndSrc := rand.NewChaCha8(seed)
	rnd := rand.New(rndSrc)
	return &Simulation{
		ts: ts,
		env: &Environment{
			src:   rndSrc,
			rnd:   rnd,
			clock: ts,
		},
		airbases: make([]Airbase, 0),
	}
}

type AirbasesOptions struct {
	IncludeRegions        []string
	ExcludeRegions        []string
	MinPerRegion          uint
	MaxPerRegion          uint
	MaxTotal              uint
	RegionProbability     prng.Ratio
	MaxAttemptsPerAirbase uint
	MetadataFactory       func(region assets.Region) map[string]any
}

type SimulationOptions struct {
	Airbases AirbasesOptions
}

func (s *Simulation) Init(config *SimulationOptions) error {
	if config == nil {
		s.airbases = s.airbases[:0]
		s.AssertInvariants()
		return nil
	}

	opts := normalizeAirbaseOptions(config.Airbases)
	airbases, err := s.generateAirbases(opts)
	if err != nil {
		return err
	}
	s.airbases = airbases
	s.AssertInvariants()
	return nil
}

func (s *Simulation) step() {
	s.env.AssertInvariants()
	s.stepTag = s.env.Rand().Uint64()
	s.ts.Tick()
}

// Clone deep copies the simulation. It pauses the
// simulation before doing so.
func (s *Simulation) Clone() *Simulation {
	ts := s.ts.Clone()
	env := s.env.Clone(ts)
	clonedAirbases := make([]Airbase, len(s.airbases))
	for i, base := range s.airbases {
		clonedAirbases[i] = base.Clone()
	}
	return &Simulation{
		ts:       ts,
		env:      env,
		stepTag:  s.stepTag,
		airbases: clonedAirbases,
	}
}

func (s *Simulation) generateAirbases(opts AirbasesOptions) ([]Airbase, error) {
	defer s.AssertInvariants()

	rng := s.env.Rand()
	include := makeRegionFilter(opts.IncludeRegions)
	exclude := makeRegionFilter(opts.ExcludeRegions)
	maxTotal := opts.MaxTotal
	result := make([]Airbase, 0)

	regions := make([]assets.Region, 0, len(assets.Regions))
	for _, region := range assets.Regions {
		if len(include) > 0 {
			if _, ok := include[strings.ToLower(region.Name)]; !ok {
				continue
			}
		}
		if len(exclude) > 0 {
			if _, ok := exclude[strings.ToLower(region.Name)]; ok {
				continue
			}
		}
		regions = append(regions, region)
	}

	sort.SliceStable(regions, func(i, j int) bool {
		return regions[i].Name < regions[j].Name
	})

	for _, region := range regions {
		if len(result) >= int(maxTotal) {
			break
		}
		if !prng.Chance(rng, opts.RegionProbability) {
			continue
		}

		polygons := make([][]geometry.Point, 0, len(region.Areas))
		areas := make([]float64, 0, len(region.Areas))
		for _, area := range region.Areas {
			if len(area) < 3 {
				continue
			}
			poly := toGeometryPolygon(area)
			polyArea := geometry.PolygonArea(poly)
			if polyArea <= 0 {
				continue
			}
			polygons = append(polygons, poly)
			areas = append(areas, polyArea)
		}
		if len(polygons) == 0 {
			continue
		}

		remaining := maxTotal - uint(len(result))
		maxPerRegion := opts.MaxPerRegion
		if maxPerRegion > remaining {
			maxPerRegion = remaining
		}
		if maxPerRegion == 0 {
			continue
		}
		minPerRegion := opts.MinPerRegion
		if minPerRegion > maxPerRegion {
			minPerRegion = maxPerRegion
		}

		count := minPerRegion
		if maxPerRegion > minPerRegion {
			count = uint(prng.RangeInclusive(rng, minPerRegion, maxPerRegion))
		}

		for i := uint(0); i < count; i++ {
			if len(result) >= int(maxTotal) {
				break
			}
			pt, err := sampleFromPolygons(rng, polygons, areas, opts.MaxAttemptsPerAirbase)
			if err != nil {
				return nil, fmt.Errorf("sample point for region %s: %w", region.Name, err)
			}
			airbase := Airbase{
				ID:       makeBaseID(rng.Uint64()),
				Location: pt,
				RegionID: region.ID,
				Region:   region.Name,
			}
			if opts.MetadataFactory != nil {
				airbase.Metadata = opts.MetadataFactory(region)
			}
			result = append(result, airbase)
		}
	}

	return result, nil
}

// Airbases returns a shallow copy of the generated airbases slice.
func (s *Simulation) Airbases() []Airbase {
	copyOf := make([]Airbase, len(s.airbases))
	copy(copyOf, s.airbases)
	return copyOf
}

func sampleFromPolygons(rng *rand.Rand, polygons [][]geometry.Point, areas []float64, maxAttempts uint) (geometry.Point, error) {
	if len(polygons) == 0 {
		return geometry.Point{}, errors.New("no polygons available")
	}

	locPolys := make([][]geometry.Point, len(polygons))
	copy(locPolys, polygons)
	locAreas := make([]float64, len(areas))
	copy(locAreas, areas)

	attempts := uint(0)
	for len(locPolys) > 0 {
		total := 0.0
		for _, area := range locAreas {
			total += area
		}
		if total == 0 {
			return geometry.Point{}, errors.New("zero polygon area")
		}

		threshold := rng.Float64() * total
		accum := 0.0
		var idx int
		for i, area := range locAreas {
			accum += area
			if threshold <= accum {
				idx = i
				break
			}
		}

		pt, err := geometry.RandomPointInPolygon(rng, locPolys[idx])
		if err == nil {
			return pt, nil
		}
		if errors.Is(err, geometry.ErrInvalidPolygon) {
			if fallback, fbErr := fallbackPointInPolygon(rng, locPolys[idx], maxAttempts); fbErr == nil {
				return fallback, nil
			}
			locPolys = append(locPolys[:idx], locPolys[idx+1:]...)
			locAreas = append(locAreas[:idx], locAreas[idx+1:]...)
			continue
		}

		attempts++
		if maxAttempts != 0 && attempts >= maxAttempts {
			return geometry.Point{}, err
		}
	}

	return geometry.Point{}, errors.New("no valid polygons")
}

func makeRegionFilter(regions []string) map[string]struct{} {
	if len(regions) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(regions))
	for _, name := range regions {
		trimmed := strings.TrimSpace(strings.ToLower(name))
		if trimmed == "" {
			continue
		}
		set[trimmed] = struct{}{}
	}
	return set
}

func normalizeAirbaseOptions(opts AirbasesOptions) AirbasesOptions {
	if opts.MinPerRegion == 0 {
		opts.MinPerRegion = 1
	}
	if opts.MaxPerRegion == 0 {
		opts.MaxPerRegion = opts.MinPerRegion
	}
	if opts.MaxPerRegion < opts.MinPerRegion {
		opts.MaxPerRegion = opts.MinPerRegion
	}
	if opts.MaxTotal == 0 {
		opts.MaxTotal = MaxNumAirbases
	}
	if opts.RegionProbability.Denominator() == 0 {
		opts.RegionProbability = prng.New(1, 1)
	}
	if opts.MaxAttemptsPerAirbase == 0 {
		opts.MaxAttemptsPerAirbase = 8
	}
	return opts
}

func makeBaseID(value uint64) BaseID {
	var id BaseID
	binary.BigEndian.PutUint64(id[:], value)
	return id
}

func toGeometryPolygon(points []assets.Point) []geometry.Point {
	poly := make([]geometry.Point, len(points))
	for i, pt := range points {
		poly[i] = geometry.Point{X: pt.X, Y: pt.Y}
	}
	return poly
}

func fallbackPointInPolygon(rng *rand.Rand, poly []geometry.Point, maxAttempts uint) (geometry.Point, error) {
	if len(poly) < 3 {
		return geometry.Point{}, geometry.ErrInvalidPolygon
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
		return geometry.Point{}, geometry.ErrInvalidPolygon
	}

	limit := maxAttempts
	if limit == 0 {
		limit = 64
	}
	for attempt := uint(0); attempt < limit; attempt++ {
		x := rng.Float64()*(maxX-minX) + minX
		y := rng.Float64()*(maxY-minY) + minY
		candidate := geometry.Point{X: x, Y: y}
		if pointInPolygonGeometry(candidate, poly) {
			return candidate, nil
		}
	}
	return geometry.Point{}, errors.New("fallback sampling exceeded attempts")
}

func pointInPolygonGeometry(p geometry.Point, poly []geometry.Point) bool {
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
