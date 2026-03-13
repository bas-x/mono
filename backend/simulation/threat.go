package simulation

import (
	"math/rand/v2"
	"time"

	"github.com/bas-x/basex/assert"
	"github.com/bas-x/basex/geometry"
	"github.com/bas-x/basex/prng"
)

const (
	mapMinX = 0.254
	mapMinY = 0.273
	mapMaxX = 345.374
	mapMaxY = 792.273
)

type ThreatID [8]byte

type Threat struct {
	ID          ThreatID
	CreatedAt   time.Time
	CreatedTick uint64
	Position    geometry.Point
}

type ThreatOptions struct {
	SpawnChance    prng.Ratio
	MaxActive      uint
	MaxActiveTicks uint64
}

type ThreatSet struct {
	pending      []Threat
	active       map[ThreatID]Threat
	nextTargetIx int
}

func NewThreatSet() *ThreatSet {
	return &ThreatSet{pending: make([]Threat, 0), active: make(map[ThreatID]Threat)}
}

func (t *ThreatSet) Clone() *ThreatSet {
	if t == nil {
		return NewThreatSet()
	}
	cp := make([]Threat, len(t.pending))
	copy(cp, t.pending)
	activeCP := make(map[ThreatID]Threat, len(t.active))
	for k, v := range t.active {
		activeCP[k] = v
	}
	return &ThreatSet{pending: cp, active: activeCP, nextTargetIx: t.nextTargetIx}
}

func (t *ThreatSet) Pending() []Threat {
	if t == nil {
		return nil
	}
	cp := make([]Threat, len(t.pending))
	copy(cp, t.pending)
	return cp
}

func (t *ThreatSet) Activate(threat Threat) {
	if t == nil {
		return
	}
	if t.active == nil {
		t.active = make(map[ThreatID]Threat)
	}
	t.active[threat.ID] = threat
}

func (t *ThreatSet) NextTarget() (Threat, bool) {
	if t == nil || len(t.pending) == 0 {
		return Threat{}, false
	}
	idx := t.nextTargetIx % len(t.pending)
	threat := t.pending[idx]
	t.nextTargetIx = (idx + 1) % len(t.pending)
	t.Activate(threat)
	return threat, true
}

func (t *ThreatSet) IsActive(id ThreatID) bool {
	if t == nil {
		return false
	}
	_, ok := t.active[id]
	return ok
}

func (t *ThreatSet) DespawnActive(tick uint64, maxAgeTicks uint64) []Threat {
	if t == nil || maxAgeTicks == 0 {
		return nil
	}
	if len(t.active) == 0 {
		return nil
	}
	despawned := make([]Threat, 0)
	for id, threat := range t.active {
		if tick-threat.CreatedTick >= maxAgeTicks {
			despawned = append(despawned, threat)
			delete(t.active, id)
			filtered := t.pending[:0]
			for _, pendingThreat := range t.pending {
				if pendingThreat.ID != id {
					filtered = append(filtered, pendingThreat)
				}
			}
			t.pending = filtered
			if len(t.pending) == 0 {
				t.nextTargetIx = 0
			} else if t.nextTargetIx >= len(t.pending) {
				t.nextTargetIx %= len(t.pending)
			}
		}
	}
	return despawned
}

func spawnEdgePoint(rng *rand.Rand, minX, minY, maxX, maxY float64) geometry.Point {
	width := maxX - minX
	height := maxY - minY
	perimeter := 2 * (width + height)
	t := rng.Float64() * perimeter
	if t < width {
		return geometry.Point{X: minX + t, Y: maxY}
	}
	if t < width+height {
		return geometry.Point{X: maxX, Y: maxY - (t - width)}
	}
	if t < 2*width+height {
		return geometry.Point{X: maxX - (t - width - height), Y: minY}
	}
	return geometry.Point{X: minX, Y: minY + (t - 2*width - height)}
}

func (t *ThreatSet) TrySpawnEdge(env *Environment, opts ThreatOptions, tick uint64) (Threat, bool) {
	if t == nil {
		return Threat{}, false
	}
	if opts.MaxActive == 0 {
		return Threat{}, false
	}
	if uint(len(t.pending)+len(t.active)) >= opts.MaxActive {
		return Threat{}, false
	}
	if opts.SpawnChance.Denominator() == 0 {
		return Threat{}, false
	}
	rng := env.Rand()
	if !prng.Chance(rng, opts.SpawnChance) {
		return Threat{}, false
	}
	pos := spawnEdgePoint(rng, mapMinX, mapMinY, mapMaxX, mapMaxY)
	threat := Threat{
		ID:          makeThreatID(rng.Uint64()),
		CreatedAt:   env.Clock().Now(),
		CreatedTick: tick,
		Position:    pos,
	}
	t.pending = append(t.pending, threat)
	return threat, true
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
	assert.InRange(t.Position.X, mapMinX, mapMaxX, "threat position x")
	assert.InRange(t.Position.Y, mapMinY, mapMaxY, "threat position y")
}
