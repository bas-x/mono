package services

import (
	"encoding/hex"
	"maps"
	"time"

	"github.com/bas-x/basex/simulation"
)

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type Airbase struct {
	ID       string         `json:"id"`
	Location Point          `json:"location"`
	RegionID string         `json:"regionId"`
	Region   string         `json:"region"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type Need struct {
	Type               string `json:"type"`
	Severity           int    `json:"severity"`
	RequiredCapability string `json:"requiredCapability"`
	Blocking           bool   `json:"blocking"`
}

type Aircraft struct {
	TailNumber string  `json:"tailNumber"`
	Needs      []Need  `json:"needs"`
	State      string  `json:"state"`
	AssignedTo *string `json:"assignedTo,omitempty"`
	Position   Point   `json:"position"`
}

type Threat struct {
	ID          string    `json:"id"`
	Position    Point     `json:"position"`
	CreatedAt   time.Time `json:"createdAt"`
	CreatedTick uint64    `json:"createdTick"`
}

type AircraftPositionSnapshot struct {
	TailNumber string `json:"tailNumber"`
	Position   Point  `json:"position"`
	State      string `json:"state"`
	Needs      []Need `json:"needs"`
}

type Assignment struct {
	Base   string `json:"base"`
	Source string `json:"source"`
}

func mapAirbase(input simulation.Airbase) Airbase {
	var metadata map[string]any
	if input.Metadata != nil {
		metadata = make(map[string]any, len(input.Metadata))
		maps.Copy(metadata, input.Metadata)
	}

	return Airbase{
		ID:       hex.EncodeToString(input.ID[:]),
		Location: Point{X: input.Location.X, Y: input.Location.Y},
		RegionID: input.RegionID,
		Region:   input.Region,
		Metadata: metadata,
	}
}

func mapAircraft(input simulation.Aircraft) Aircraft {
	needs := make([]Need, len(input.Needs))
	for i, need := range input.Needs {
		needs[i] = Need{
			Type:               string(need.Type),
			Severity:           need.Severity,
			RequiredCapability: string(need.RequiredCapability),
			Blocking:           need.Blocking,
		}
	}

	var assignedTo *string
	if input.HasAssignment {
		baseID := hex.EncodeToString(input.AssignedBase[:])
		assignedTo = &baseID
	}

	return Aircraft{
		TailNumber: hex.EncodeToString(input.TailNumber[:]),
		Needs:      needs,
		State:      input.State.Name(),
		AssignedTo: assignedTo,
		Position:   Point{X: input.Position.X, Y: input.Position.Y},
	}
}

func mapThreat(input simulation.Threat) Threat {
	return Threat{
		ID:          hex.EncodeToString(input.ID[:]),
		Position:    Point{X: input.Position.X, Y: input.Position.Y},
		CreatedAt:   input.CreatedAt,
		CreatedTick: input.CreatedTick,
	}
}
