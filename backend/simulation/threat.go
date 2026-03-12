package simulation

import (
	"sort"
	"time"

	"github.com/bas-x/basex/assert"
	"github.com/bas-x/basex/assets"
	"github.com/bas-x/basex/prng"
)

type ThreatID [8]byte

type Threat struct {
	ID          ThreatID
	RegionID    string
	Region      string
	CreatedAt   time.Time
	CreatedTick uint64
}

type ThreatOptions struct {
	SpawnChance prng.Ratio
	MaxActive   uint
}

type ThreatSet struct {
	pending []Threat
}

func NewThreatSet() *ThreatSet {
	return &ThreatSet{pending: make([]Threat, 0)}
}

func (t *ThreatSet) Clone() *ThreatSet {
	if t == nil {
		return NewThreatSet()
	}
	cp := make([]Threat, len(t.pending))
	copy(cp, t.pending)
	return &ThreatSet{pending: cp}
}

func (t *ThreatSet) Pending() []Threat {
	if t == nil {
		return nil
	}
	cp := make([]Threat, len(t.pending))
	copy(cp, t.pending)
	return cp
}

func (t *ThreatSet) TrySpawn(env *Environment, bases []Airbase, opts ThreatOptions, tick uint64) (Threat, bool) {
	return t.TrySpawnFromRegions(env, threatRegionsFromBases(bases), opts, tick)
}

func (t *ThreatSet) TrySpawnFromRegions(env *Environment, regions []threatRegion, opts ThreatOptions, tick uint64) (Threat, bool) {
	if t == nil {
		return Threat{}, false
	}
	if opts.MaxActive == 0 {
		return Threat{}, false
	}
	if uint(len(t.pending)) >= opts.MaxActive {
		return Threat{}, false
	}
	if opts.SpawnChance.Denominator() == 0 {
		return Threat{}, false
	}
	rng := env.Rand()
	if !prng.Chance(rng, opts.SpawnChance) {
		return Threat{}, false
	}
	if len(regions) == 0 {
		return Threat{}, false
	}
	idx := int(prng.RangeInclusive(rng, 0, uint(len(regions)-1)))
	region := regions[idx]
	threat := Threat{
		ID:          makeThreatID(rng.Uint64()),
		RegionID:    region.RegionID,
		Region:      region.Region,
		CreatedAt:   env.Clock().Now(),
		CreatedTick: tick,
	}
	t.pending = append(t.pending, threat)
	return threat, true
}

func (t *ThreatSet) ClaimNext() (Threat, bool) {
	if t == nil || len(t.pending) == 0 {
		return Threat{}, false
	}
	threat := t.pending[0]
	t.pending = t.pending[1:]
	return threat, true
}

type threatRegion struct {
	RegionID string
	Region   string
}

func threatRegionsFromBases(bases []Airbase) []threatRegion {
	if len(bases) == 0 {
		return nil
	}
	seen := make(map[string]threatRegion, len(bases))
	for _, base := range bases {
		key := base.RegionID + "|" + base.Region
		seen[key] = threatRegion{RegionID: base.RegionID, Region: base.Region}
	}
	regions := make([]threatRegion, 0, len(seen))
	for _, region := range seen {
		regions = append(regions, region)
	}
	sort.SliceStable(regions, func(i, j int) bool {
		if regions[i].Region == regions[j].Region {
			return regions[i].RegionID < regions[j].RegionID
		}
		return regions[i].Region < regions[j].Region
	})
	return regions
}

func threatRegionsFromNames(names []string) []threatRegion {
	if len(names) == 0 {
		return nil
	}
	regions := make([]threatRegion, 0, len(names))
	seen := make(map[string]struct{}, len(names))
	for _, name := range names {
		if name == "" {
			continue
		}
		bounds, ok := assets.GetRegionBounds(name)
		if !ok {
			continue
		}
		if _, exists := seen[bounds.ID]; exists {
			continue
		}
		seen[bounds.ID] = struct{}{}
		regions = append(regions, threatRegion{RegionID: bounds.ID, Region: bounds.Name})
	}
	sort.SliceStable(regions, func(i, j int) bool {
		if regions[i].Region == regions[j].Region {
			return regions[i].RegionID < regions[j].RegionID
		}
		return regions[i].Region < regions[j].Region
	})
	return regions
}

func threatRegionsAll() []threatRegion {
	regions := make([]threatRegion, 0, len(assets.BoundsData.Regions))
	for _, region := range assets.BoundsData.Regions {
		regions = append(regions, threatRegion{RegionID: region.ID, Region: region.Name})
	}
	sort.SliceStable(regions, func(i, j int) bool {
		if regions[i].Region == regions[j].Region {
			return regions[i].RegionID < regions[j].RegionID
		}
		return regions[i].Region < regions[j].Region
	})
	return regions
}

func makeThreatID(value uint64) ThreatID {
	var id ThreatID
	for i := 7; i >= 0; i-- {
		id[i] = byte(value)
		value >>= 8
	}
	return id
}

func (t Threat) AssertInvariants() {
	assert.True(t.RegionID != "", "threat region id", t.RegionID)
	assert.True(t.Region != "", "threat region", t.Region)
}
