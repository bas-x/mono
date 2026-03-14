package simulation

import (
	"encoding/binary"
	"sort"
	"strconv"
	"strings"

	"github.com/bas-x/basex/assert"
	"github.com/bas-x/basex/assets"
	"github.com/bas-x/basex/geometry"
	"github.com/bas-x/basex/prng"
)

const (
	MaxNumAirbases = 256
)

type Constellation struct {
	airbases []Airbase
}

type ConstellationOptions struct {
	IncludeRegions        []string
	ExcludeRegions        []string
	MinPerRegion          uint
	MaxPerRegion          uint
	MaxTotal              uint
	RegionProbability     prng.Ratio
	MaxAttemptsPerAirbase uint
	MetadataFactory       func(region assets.Region) map[string]any
}

func NewConstellation() *Constellation {
	return &Constellation{
		airbases: make([]Airbase, 0),
	}
}

func (c *Constellation) AssertInvariants() {
	assert.NotNil(c, "constellation")
	assert.InRange(len(c.airbases), 0, MaxNumAirbases, "airbases length")

	seen := make(map[BaseID]struct{}, len(c.airbases))
	for _, base := range c.airbases {
		if _, ok := seen[base.ID]; ok {
			panic("duplicate airbase id")
		}
		seen[base.ID] = struct{}{}
	}
}

func (c *Constellation) Init(env *Environment, opts *ConstellationOptions) error {
	assert.NotNil(env, "environment")
	env.AssertInvariants()

	if c.airbases == nil {
		c.airbases = make([]Airbase, 0)
	}

	if opts == nil {
		c.airbases = c.airbases[:0]
		c.AssertInvariants()
		return nil
	}

	normalized := normalizeAirbaseOptions(*opts)
	bases, err := c.generateAirbases(env, normalized)
	if err != nil {
		return err
	}
	c.airbases = bases
	c.AssertInvariants()
	return nil
}

func (c *Constellation) Clone() *Constellation {
	cloned := make([]Airbase, len(c.airbases))
	for i, base := range c.airbases {
		cloned[i] = base.Clone()
	}
	return &Constellation{airbases: cloned}
}

func (c *Constellation) Airbases() []Airbase {
	copyOf := make([]Airbase, len(c.airbases))
	copy(copyOf, c.airbases)
	return copyOf
}

func (c *Constellation) generateAirbases(env *Environment, opts ConstellationOptions) ([]Airbase, error) {
	rng := env.Rand()
	include := makeRegionFilter(opts.IncludeRegions)
	exclude := makeRegionFilter(opts.ExcludeRegions)
	maxTotal := opts.MaxTotal
	airbases := make([]Airbase, 0)

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
		if len(airbases) >= int(maxTotal) {
			break
		}
		if !prng.Chance(rng, opts.RegionProbability) {
			continue
		}

		polygons := make([][]geometry.Point, 0, len(region.Areas))
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
		}
		if len(polygons) == 0 {
			continue
		}

		remaining := maxTotal - uint(len(airbases))
		maxPerRegion := min(opts.MaxPerRegion, remaining)
		if maxPerRegion == 0 {
			continue
		}

		minPerRegion := min(opts.MinPerRegion, maxPerRegion)

		count := minPerRegion
		if maxPerRegion > minPerRegion {
			count = uint(prng.RangeInclusive(rng, minPerRegion, maxPerRegion))
		}

		for i := uint(0); i < count; i++ {
			if len(airbases) >= int(maxTotal) {
				break
			}

			pt, ok := geometry.SampleFromPolygons(rng, polygons, opts.MaxAttemptsPerAirbase)
			if !ok {
				continue
			}
			nameIndex := len(airbases)
			name := assets.AirbaseNames[nameIndex%len(assets.AirbaseNames)]
			if nameIndex >= len(assets.AirbaseNames) {
				name = name + " " + strconv.Itoa(nameIndex/len(assets.AirbaseNames)+1)
			}
			airbase := Airbase{
				ID:           makeBaseID(rng.Uint64()),
				Name:         name,
				Location:     pt,
				RegionID:     region.ID,
				Region:       region.Name,
				Capabilities: defaultAirbaseCapabilities(),
			}
			if opts.MetadataFactory != nil {
				airbase.Metadata = opts.MetadataFactory(region)
			}
			airbases = append(airbases, airbase)
		}
	}

	return airbases, nil
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

func normalizeAirbaseOptions(opts ConstellationOptions) ConstellationOptions {
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
