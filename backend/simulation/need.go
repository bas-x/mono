package simulation

import (
	"github.com/bas-x/basex/assert"
)

type NeedType string

const (
	NeedFuel                 NeedType = "fuel"
	NeedCharge               NeedType = "charge"
	NeedMunitions            NeedType = "munitions"
	NeedRepairs              NeedType = "repairs"
	NeedMaintenance          NeedType = "maintenance"
	NeedMissionConfiguration NeedType = "mission_configuration"
	NeedCrewSupport          NeedType = "crew_support"
	NeedEmergency            NeedType = "emergency"
	NeedWeatherConstraint    NeedType = "weather_constraint"
	NeedGroundSupport        NeedType = "ground_support"
	NeedProtection           NeedType = "protection"
)

var (
	AllNeedTypes = []NeedType{
		NeedFuel,
		NeedCharge,
		NeedMunitions,
		NeedRepairs,
		NeedMaintenance,
		NeedMissionConfiguration,
		NeedCrewSupport,
		NeedEmergency,
		NeedWeatherConstraint,
		NeedGroundSupport,
		NeedProtection,
	}
)

type Need struct {
	Type               NeedType
	Severity           int // 0-100
	RequiredCapability NeedType
	Blocking           bool
}

// NeedTypeIndex returns the stable index of a NeedType within AllNeedTypes.
func NeedTypeIndex(t NeedType) (int, bool) {
	switch t {
	case NeedFuel:
		return 0, true
	case NeedCharge:
		return 1, true
	case NeedMunitions:
		return 2, true
	case NeedRepairs:
		return 3, true
	case NeedMaintenance:
		return 4, true
	case NeedMissionConfiguration:
		return 5, true
	case NeedCrewSupport:
		return 6, true
	case NeedEmergency:
		return 7, true
	case NeedWeatherConstraint:
		return 8, true
	case NeedGroundSupport:
		return 9, true
	case NeedProtection:
		return 10, true
	default:
		return -1, false
	}
}

// IsValidNeedType reports whether the provided NeedType is known to the simulation.
func IsValidNeedType(t NeedType) bool {
	_, ok := NeedTypeIndex(t)
	return ok
}

// AssertInvariants validates the internal consistency of the Need.
func (n Need) AssertInvariants() {
	assert.True(IsValidNeedType(n.Type), "need type", n.Type)
	assert.InRange(n.Severity, 0, 100, "need severity", n)
	if n.RequiredCapability != "" {
		assert.True(IsValidNeedType(n.RequiredCapability), "need required capability", n.RequiredCapability)
	}
}

// Clone returns a deep copy of the Need.
func (n Need) Clone() Need {
	return Need{
		Type:               n.Type,
		Severity:           n.Severity,
		RequiredCapability: n.RequiredCapability,
		Blocking:           n.Blocking,
	}
}
