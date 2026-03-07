package assets

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

// Point describes a 2D coordinate in SVG space.
type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// Region represents a geographic region extracted from the Sweden SVG map.
type Region struct {
	ID    string    `json:"id"`
	Name  string    `json:"name"`
	Areas [][]Point `json:"areas"`
}

// Bounds captures the bounding rectangle for a region or collection of regions.
type Bounds struct {
	Min    Point   `json:"min"`
	Max    Point   `json:"max"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// RegionBounds is the per-region entry contained in bounds.json.
type RegionBounds struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Bounds
}

// BoundsFile mirrors the top-level structure of bounds.json.
type BoundsFile struct {
	Overall Bounds         `json:"overall"`
	Regions []RegionBounds `json:"regions"`
}

//go:embed sweden.json
var swedenJSON []byte

//go:embed bounds.json
var boundsJSON []byte

// Regions provides the parsed contents of sweden.json.
var Regions []Region

// BoundsData provides the parsed contents of bounds.json.
var BoundsData BoundsFile

// RegionBoundsByName maps a region name to its bounding box entry.
var RegionBoundsByName map[string]RegionBounds

func init() {
	if err := json.Unmarshal(swedenJSON, &Regions); err != nil {
		panic(fmt.Errorf("assets: failed to decode sweden.json: %w", err))
	}

	if err := json.Unmarshal(boundsJSON, &BoundsData); err != nil {
		panic(fmt.Errorf("assets: failed to decode bounds.json: %w", err))
	}

	RegionBoundsByName = make(map[string]RegionBounds, len(BoundsData.Regions))
	for _, entry := range BoundsData.Regions {
		RegionBoundsByName[entry.Name] = entry
	}
}

// GetRegionBounds returns the bounding box for the provided region name.
func GetRegionBounds(name string) (RegionBounds, bool) {
	entry, ok := RegionBoundsByName[name]
	return entry, ok
}
