package simulation

import "time"

type NeedPhase int

const (
	NeedPhaseOutbound NeedPhase = iota
	NeedPhaseEngaged
	NeedPhaseServicing
)

type PhaseDurations struct {
	Outbound        time.Duration
	Engaged         time.Duration
	InboundDecision time.Duration
	CommitApproach  time.Duration
	Servicing       time.Duration
}

type NeedRateModel struct {
	OutboundMilliPerHour  int64
	EngagedMilliPerHour   int64
	ServicingMilliPerHour int64
	VariancePermille      int64
}

func (m NeedRateModel) RateForPhase(phase NeedPhase) int64 {
	switch phase {
	case NeedPhaseOutbound:
		return m.OutboundMilliPerHour
	case NeedPhaseEngaged:
		return m.EngagedMilliPerHour
	case NeedPhaseServicing:
		return m.ServicingMilliPerHour
	default:
		return 0
	}
}

type LifecycleModel struct {
	Durations       PhaseDurations
	ReturnThreshold int
	NeedRates       map[NeedType]NeedRateModel
}

func DefaultLifecycleModel() LifecycleModel {
	return LifecycleModel{
		Durations: PhaseDurations{
			Outbound:        25 * time.Minute,
			Engaged:         35 * time.Minute,
			InboundDecision: 8 * time.Minute,
			CommitApproach:  6 * time.Minute,
			Servicing:       75 * time.Minute,
		},
		ReturnThreshold: 85,
		NeedRates: map[NeedType]NeedRateModel{
			NeedFuel:                 {OutboundMilliPerHour: 2600, EngagedMilliPerHour: 4200, ServicingMilliPerHour: 12000, VariancePermille: 120},
			NeedCharge:               {OutboundMilliPerHour: 1800, EngagedMilliPerHour: 2400, ServicingMilliPerHour: 10000, VariancePermille: 100},
			NeedMunitions:            {OutboundMilliPerHour: 1100, EngagedMilliPerHour: 3600, ServicingMilliPerHour: 9000, VariancePermille: 140},
			NeedRepairs:              {OutboundMilliPerHour: 900, EngagedMilliPerHour: 1700, ServicingMilliPerHour: 7000, VariancePermille: 160},
			NeedMaintenance:          {OutboundMilliPerHour: 700, EngagedMilliPerHour: 1200, ServicingMilliPerHour: 6500, VariancePermille: 130},
			NeedMissionConfiguration: {OutboundMilliPerHour: 500, EngagedMilliPerHour: 900, ServicingMilliPerHour: 8500, VariancePermille: 120},
			NeedCrewSupport:          {OutboundMilliPerHour: 650, EngagedMilliPerHour: 900, ServicingMilliPerHour: 8000, VariancePermille: 100},
			NeedEmergency:            {OutboundMilliPerHour: 400, EngagedMilliPerHour: 800, ServicingMilliPerHour: 9500, VariancePermille: 80},
			NeedWeatherConstraint:    {OutboundMilliPerHour: 250, EngagedMilliPerHour: 450, ServicingMilliPerHour: 5000, VariancePermille: 90},
			NeedGroundSupport:        {OutboundMilliPerHour: 500, EngagedMilliPerHour: 700, ServicingMilliPerHour: 7800, VariancePermille: 110},
			NeedProtection:           {OutboundMilliPerHour: 350, EngagedMilliPerHour: 650, ServicingMilliPerHour: 6200, VariancePermille: 90},
		},
	}
}

func (m LifecycleModel) NeedModel(t NeedType) NeedRateModel {
	model, ok := m.NeedRates[t]
	if ok {
		return model
	}
	return NeedRateModel{OutboundMilliPerHour: 600, EngagedMilliPerHour: 900, ServicingMilliPerHour: 7000, VariancePermille: 100}
}
