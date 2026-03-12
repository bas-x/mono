package simulation

import (
	"maps"

	"github.com/bas-x/basex/geometry"
)

type BaseID [8]byte

type Airbase struct {
	ID           BaseID
	Location     geometry.Point
	RegionID     string
	Region       string
	Capabilities map[NeedType]AirbaseCapability
	Metadata     map[string]any
}

type AirbaseCapability struct {
	RecoveryMultiplierPermille int64
}

// Clone returns a deep copy of the airbase.
func (a Airbase) Clone() Airbase {
	var meta map[string]any
	if a.Metadata != nil {
		meta = make(map[string]any, len(a.Metadata))
		maps.Copy(meta, a.Metadata)
	}
	return Airbase{
		ID:           a.ID,
		Location:     a.Location,
		RegionID:     a.RegionID,
		Region:       a.Region,
		Capabilities: cloneCapabilities(a.Capabilities),
		Metadata:     meta,
	}
}

func (a Airbase) RecoveryMultiplier(cap NeedType) int64 {
	if capability, ok := a.Capabilities[cap]; ok {
		if capability.RecoveryMultiplierPermille > 0 {
			return capability.RecoveryMultiplierPermille
		}
	}
	return 1000
}

func defaultAirbaseCapabilities() map[NeedType]AirbaseCapability {
	return map[NeedType]AirbaseCapability{
		NeedFuel:                 {RecoveryMultiplierPermille: 1300},
		NeedCharge:               {RecoveryMultiplierPermille: 1200},
		NeedMunitions:            {RecoveryMultiplierPermille: 1000},
		NeedRepairs:              {RecoveryMultiplierPermille: 800},
		NeedMaintenance:          {RecoveryMultiplierPermille: 900},
		NeedMissionConfiguration: {RecoveryMultiplierPermille: 1050},
		NeedCrewSupport:          {RecoveryMultiplierPermille: 1100},
		NeedEmergency:            {RecoveryMultiplierPermille: 1200},
		NeedWeatherConstraint:    {RecoveryMultiplierPermille: 700},
		NeedGroundSupport:        {RecoveryMultiplierPermille: 950},
		NeedProtection:           {RecoveryMultiplierPermille: 850},
	}
}

func cloneCapabilities(input map[NeedType]AirbaseCapability) map[NeedType]AirbaseCapability {
	if input == nil {
		return nil
	}
	cloned := make(map[NeedType]AirbaseCapability, len(input))
	maps.Copy(cloned, input)
	return cloned
}
