package simulation

import (
	"maps"

	"github.com/bas-x/basex/geometry"
)

type BaseID [8]byte

type Airbase struct {
	ID       BaseID
	Location geometry.Point
	RegionID string
	Region   string
	Metadata map[string]any
}

// Clone returns a deep copy of the airbase.
func (a Airbase) Clone() Airbase {
	var meta map[string]any
	if a.Metadata != nil {
		meta = make(map[string]any, len(a.Metadata))
		maps.Copy(meta, a.Metadata)
	}
	return Airbase{
		ID:       a.ID,
		Location: a.Location,
		RegionID: a.RegionID,
		Region:   a.Region,
		Metadata: meta,
	}
}
