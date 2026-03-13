package simulation

import (
	"testing"

	"github.com/bas-x/basex/geometry"
)

func TestAircraftCloneNewFields(t *testing.T) {
	// Create a threat to test ClaimedThreat cloning
	threat := &Threat{
		ID:          [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
		Position:    geometry.Point{X: 150.0, Y: 250.0},
		CreatedTick: 42,
	}

	// Create an aircraft with non-zero values for all 5 new fields
	original := &Aircraft{
		TailNumber:     [8]byte{10, 20, 30, 40, 50, 60, 70, 80},
		State:          &OutboundState{},
		Position:       geometry.Point{X: 100.5, Y: 200.5},
		Speed:          95.5,
		OrbitAngle:     1.57,
		ClaimedThreat:  threat,
		ThreatCentroid: geometry.Point{X: 300.0, Y: 400.0},
	}

	// Clone the aircraft
	cloned := original.Clone()

	// Assert all 5 new fields are copied correctly
	if cloned.Position != original.Position {
		t.Errorf("Position not cloned: got %v, want %v", cloned.Position, original.Position)
	}
	if cloned.Speed != original.Speed {
		t.Errorf("Speed not cloned: got %v, want %v", cloned.Speed, original.Speed)
	}
	if cloned.OrbitAngle != original.OrbitAngle {
		t.Errorf("OrbitAngle not cloned: got %v, want %v", cloned.OrbitAngle, original.OrbitAngle)
	}
	if cloned.ThreatCentroid != original.ThreatCentroid {
		t.Errorf("ThreatCentroid not cloned: got %v, want %v", cloned.ThreatCentroid, original.ThreatCentroid)
	}

	// Assert ClaimedThreat is cloned (not the same pointer)
	if cloned.ClaimedThreat == nil {
		t.Error("ClaimedThreat is nil after clone")
	}
	if cloned.ClaimedThreat == original.ClaimedThreat {
		t.Error("ClaimedThreat is the same pointer after clone (should be a copy)")
	}
	if cloned.ClaimedThreat.ID != original.ClaimedThreat.ID {
		t.Errorf("ClaimedThreat.ID not cloned: got %v, want %v", cloned.ClaimedThreat.ID, original.ClaimedThreat.ID)
	}
	if cloned.ClaimedThreat.Position != original.ClaimedThreat.Position {
		t.Errorf("ClaimedThreat.Position not cloned: got %v, want %v", cloned.ClaimedThreat.Position, original.ClaimedThreat.Position)
	}
	if cloned.ClaimedThreat.CreatedTick != original.ClaimedThreat.CreatedTick {
		t.Errorf("ClaimedThreat.CreatedTick not cloned: got %v, want %v", cloned.ClaimedThreat.CreatedTick, original.ClaimedThreat.CreatedTick)
	}
}

func TestAircraftSpeed(t *testing.T) {
	// Test determinism: same tail number should produce same speed
	tail := [8]byte{1, 2, 3, 4, 5, 6, 7, 8}
	speed1 := aircraftSpeed(tail)
	speed2 := aircraftSpeed(tail)

	if speed1 != speed2 {
		t.Errorf("aircraftSpeed is not deterministic: got %v and %v for same tail", speed1, speed2)
	}

	if speed1 < 1.0 || speed1 > 10.5 {
		t.Errorf("aircraftSpeed out of range [1.0, 10.5]: got %v", speed1)
	}

	// Test with different tail numbers produce different speeds (likely, but not guaranteed)
	tail2 := [8]byte{8, 7, 6, 5, 4, 3, 2, 1}
	speed3 := aircraftSpeed(tail2)
	if speed3 < 1.0 || speed3 > 10.5 {
		t.Errorf("aircraftSpeed out of range [1.0, 10.5] for tail2: got %v", speed3)
	}
}
